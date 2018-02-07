package ether

import (
	"encoding/json"
	"ethereum-front/abi"
	"ethereum-front/abi/bind"
	"ethereum-front/templates"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/compiler"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"math/big"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
)

//Object satisfying the interface should be able to:
type ReadWriterEth interface {
	Transact() (string, error)
	Call() (string, error)
	Deploy() (string, string, error)
	Info() (*Info, error)
	ParseInput() ([]interface{}, error)
	ParseOutput([]interface{}) (string, error)
}

type Info struct {
	Address         string `json:"address"`
	Balance         string `json:"balance"`
	EthBalance      string `json:"eth_balance"`
	Container       string `json:"sol_file"`
	Contract        string `json:"contract"`
	ContractAddress string `json:"contract_address"`
}

type EthWorker struct {
	Container       string
	Contract        string
	Endpoint        string
	Key             string
	ContractAddress string
	FormValues      url.Values
	New             bool
}

type ContractContainers struct {
	ContainerNames []string
	Containers     map[string]*ContractContainer
}

type ContractContainer struct {
	ContainerName string
	ContractNames []string
	Contracts     map[string]*Contract
}

type Contract struct {
	Name     string
	Abi      abi.ABI
	AbiJson  string
	Bin      string
	SortKeys []string

	OutputsInterfaces map[string][]interface{}
	InputsInterfaces  map[string][]interface{}
}

var (
	Client     bind.ContractBackend
	Containers *ContractContainers
	GasLimit   *big.Int
)

func NewEthWorker(
	container,
	contract,
	endpoint,
	key,
	contractAddress string,
	formValues url.Values,
) *EthWorker {
	return &EthWorker{
		Container:       container,
		Contract:        contract,
		Endpoint:        endpoint,
		Key:             key,
		ContractAddress: contractAddress,
		FormValues:      formValues,
	}
}

func (w *EthWorker) Transact() (string, error) {

	inputs, err := w.ParseInput()
	if err != nil {
		return "", errors.Wrap(err, "parse input")
	}

	pk := strings.TrimPrefix(w.Key, "0x")

	key, err := crypto.HexToECDSA(pk)
	if err != nil {
		return "", errors.Wrap(err, "hex to ECDSA")
	}
	auth := bind.NewKeyedTransactor(key)

	if !common.IsHexAddress(w.ContractAddress) {
		return "", errors.New("New Address From Hex")
	}

	addr := common.HexToAddress(w.ContractAddress)

	contract := bind.NewBoundContract(
		addr,
		Containers.Containers[w.Container].Contracts[w.Contract].Abi,
		Client,
		Client,
	)

	gasprice, err := Client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "suggest gas price")
	}

	opt := &bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasPrice: gasprice,
		GasLimit: GasLimit,
		Value:    auth.Value,
	}

	tr, err := contract.Transact(opt, w.Endpoint, inputs...)
	if err != nil {
		return "", errors.Wrap(err, "transact")
	}

	var receipt *types.Receipt

	switch v := Client.(type) {
	case *backends.SimulatedBackend:
		v.Commit()
		receipt, err = v.TransactionReceipt(context.Background(), tr.Hash())
		if err != nil {
			return "", errors.Wrap(err, "transaction receipt")
		}

	case *ethclient.Client:
		receipt, err = bind.WaitMined(context.Background(), v, tr)
		if err != nil {
			return "", errors.Wrap(err, "transaction receipt")
		}
	}

	if err != nil {
		return "", errors.Errorf("error transact %s: %s",
			tr.Hash().String(),
			err.Error(),
		)
	}

	responce := fmt.Sprintf(templates.WriteResult,
		tr.Nonce(),
		auth.From.String(),
		tr.To().String(),
		tr.Value().String(),
		tr.GasPrice().String(),
		receipt.GasUsed.String(),
		new(big.Int).Mul(receipt.GasUsed, tr.GasPrice()),
		receipt.Status,
		receipt.TxHash.String(),
	)

	return responce, nil
}

func (w *EthWorker) Call() (string, error) {

	inputs, err := w.ParseInput()
	if err != nil {
		return "", errors.Wrap(err, "parse input")
	}

	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	contract := bind.NewBoundContract(
		common.HexToAddress(w.ContractAddress),
		Containers.Containers[w.Container].Contracts[w.Contract].Abi,
		Client,
		Client,
	)

	opt := &bind.CallOpts{
		Pending: true,
		From:    auth.From,
	}

	outputs := Containers.Containers[w.Container].Contracts[w.Contract].OutputsInterfaces[w.Endpoint]

	if err := contract.Call(
		opt,
		&outputs,
		w.Endpoint,
		inputs...,
	); err != nil {
		return "", errors.Wrap(err, "call contract")
	}

	result, err := w.ParseOutput(outputs)
	if err != nil {
		return "", errors.Wrap(err, "parse output")
	}

	return result, err
}

func (w *EthWorker) Deploy() (string, string, error) {
	inputs, err := w.ParseInput()
	if err != nil {
		return "", "", errors.Wrap(err, "parse input")
	}

	pk := strings.TrimPrefix(w.Key, "0x")

	key, err := crypto.HexToECDSA(pk)
	if err != nil {
		return "", "", errors.Wrap(err, "hex to ECDSA")
	}
	auth := bind.NewKeyedTransactor(key)

	current_bytecode := Containers.Containers[w.Container].Contracts[w.Contract].Bin
	current_abi := Containers.Containers[w.Container].Contracts[w.Contract].Abi

	addr, tr, _, err := bind.DeployContract(auth, current_abi, common.FromHex(current_bytecode), Client, inputs...)
	if err != nil {
		log.Printf("error %s", err.Error())
		return "", "", errors.Wrap(err, "deploy contract")
	}
	var receipt *types.Receipt

	switch v := Client.(type) {
	case *backends.SimulatedBackend:
		v.Commit()
		receipt, err = v.TransactionReceipt(context.Background(), tr.Hash())
		if err != nil {
			return "", "", errors.Wrap(err, "transaction receipt")
		}

	case *ethclient.Client:
		receipt, err = bind.WaitMined(context.Background(), v, tr)
		if err != nil {
			return "", "", errors.Wrap(err, "transaction receipt")
		}
	}

	if err != nil {
		return "", "", errors.Errorf("error transact %s: %s",
			tr.Hash().String(),
			err.Error(),
		)
	}

	responce := fmt.Sprintf(templates.DeployResult,
		tr.Nonce(),
		auth.From.String(),
		addr.String(),
		tr.GasPrice().String(),
		receipt.GasUsed.String(),
		new(big.Int).Mul(receipt.GasUsed, tr.GasPrice()).String(),
		receipt.Status,
		receipt.TxHash.String(),
	)

	return responce, addr.String(), nil
}

func (w *EthWorker) Info() (*Info, error) {
	result := new(Info)

	pk := strings.TrimPrefix(w.Key, "0x")

	key, err := crypto.HexToECDSA(pk)
	if err != nil {
		return nil, errors.Wrap(err, "hex to ECDSA")
	}
	keyAddr := crypto.PubkeyToAddress(key.PublicKey)

	switch v := Client.(type) {
	case *ethclient.Client:
		balance, err := v.BalanceAt(
			context.Background(),
			keyAddr,
			nil,
		)
		if err != nil {
			return nil, errors.Wrap(err, "get balance")
		}
		result.Balance = balance.String()
	case *backends.SimulatedBackend:
		balance, err := v.BalanceAt(
			context.Background(),
			keyAddr,
			nil,
		)
		if err != nil {
			return nil, errors.Wrap(err, "get balance")
		}
		result.Balance = balance.String()
	}

	bal, _, err := new(big.Float).Parse(result.Balance, 10)
	if err != nil {
		return nil, errors.Wrap(err, "parce balance")
	}

	eth, _, _ := new(big.Float).Parse("1000000000000000000", 10)

	eth_bal := new(big.Float).Quo(bal, eth)

	result.EthBalance = eth_bal.Text('f', 10)

	result.Address = keyAddr.String()

	result.Container = w.Container

	result.Contract = w.Contract

	result.ContractAddress = w.ContractAddress

	return result, err
}

func (w *EthWorker) ParseInput() ([]interface{}, error) {

	if w.New && len(Containers.Containers[w.Container].Contracts[w.Contract].Abi.Constructor.Inputs) == 0 {
		return nil, nil
	}

	if !w.New && len(Containers.Containers[w.Container].Contracts[w.Contract].Abi.Methods[w.Endpoint].Inputs) == 0 {
		return nil, nil
	}

	inputsMap := make(map[int]string)
	var inputsArray []int
	var inputsSort []string

	for k, v := range w.FormValues {
		if k == "endpoint" {
			continue
		}

		if len(v) != 1 {
			return nil, errors.Errorf("incorrect %s field", k)
		}

		i, err := strconv.Atoi(k)
		if err != nil {
			continue
			//return nil, errors.Wrap(err, "incorrect inputs: strconv.Atoi")
		}
		inputsMap[i] = v[0]
	}

	if Containers.Containers[w.Container] == nil || Containers.Containers[w.Container].Contracts[w.Contract] == nil {
		return nil, errors.New("input values incorrect")
	}

	if !w.New && len(Containers.Containers[w.Container].Contracts[w.Contract].Abi.Methods[w.Endpoint].Inputs) != 0 && Containers.Containers[w.Container].Contracts[w.Contract].InputsInterfaces[w.Endpoint] == nil {
		return nil, errors.New("input values incorrect")
	}

	var inputs_args []abi.Argument

	if w.New {
		inputs_args = Containers.Containers[w.Container].Contracts[w.Contract].Abi.Constructor.Inputs
	} else {
		inputs_args = Containers.Containers[w.Container].Contracts[w.Contract].Abi.Methods[w.Endpoint].Inputs
	}

	if len(inputsMap) != len(inputs_args) {
		return nil, errors.New("len inputs_args != inputsMap: incorrect inputs")
	}

	for k := range inputsMap {
		inputsArray = append(inputsArray, k)
	}
	sort.Ints(inputsArray)

	for k := range inputsArray {
		inputsSort = append(inputsSort, inputsMap[k])
	}

	var inputs_interfaces []interface{}

	for i := 0; i < len(inputs_args); i++ {

		arg_value := inputsMap[i]

		switch inputs_args[i].Type.Type.String() {
		case "bool":
			var result bool
			result, err := strconv.ParseBool(arg_value)
			if err != nil {
				return nil, errors.New("incorrect inputs")
			}
			inputs_interfaces = append(inputs_interfaces, result)

		case "[]bool":
			var result []bool

			result_array := strings.Split(arg_value, ",")

			for _, bool_value := range result_array {
				item, err := strconv.ParseBool(bool_value)
				if err != nil {
					return nil, errors.Wrap(err, "incorrect inputs")
				}
				result = append(result, item)
			}
			inputs_interfaces = append(inputs_interfaces, result)

		case "string":
			inputs_interfaces = append(inputs_interfaces, arg_value)
		case "[]string":
			result_array := strings.Split(arg_value, ",") //TODO: NEED REF
			inputs_interfaces = append(inputs_interfaces, result_array)
		case "[]byte":
			inputs_interfaces = append(inputs_interfaces, []byte(arg_value))
		case "[][]byte":
			var result [][]byte

			result_array := strings.Split(arg_value, ",")

			for _, byte_value := range result_array {
				result = append(result, []byte(byte_value))
			}
			inputs_interfaces = append(inputs_interfaces, result)

		case "common.Address":
			if !common.IsHexAddress(arg_value) {
				return nil, errors.New("incorrect inputs: arg_value is not address")
			}

			inputs_interfaces = append(inputs_interfaces, common.HexToAddress(arg_value))
		case "[]common.Address":
			var result []common.Address

			result_array := strings.Split(arg_value, ",")

			for _, addr_value := range result_array {

				if !common.IsHexAddress(arg_value) {
					return nil, errors.New("incorrect inputs: arg_value is not address")
				}

				addr := common.HexToAddress(addr_value)

				result = append(result, addr)
			}
			inputs_interfaces = append(inputs_interfaces, result)
		case "common.Hash":
			if !common.IsHex(arg_value) {
				return nil, errors.New("incorrect inputs: arg_value is not hex")
			}

			inputs_interfaces = append(inputs_interfaces, common.HexToHash(arg_value))

		case "[]common.Hash":
			var result []common.Hash

			result_array := strings.Split(arg_value, ",")

			for _, addr_value := range result_array {

				if !common.IsHex(arg_value) {
					return nil, errors.New("incorrect inputs: arg_value is not hex")
				}

				hash := common.HexToHash(addr_value)
				result = append(result, hash)
			}
			inputs_interfaces = append(inputs_interfaces, result)
		case "int8":
			i, err := strconv.ParseInt(arg_value, 10, 8)
			if err != nil {
				return nil, errors.New("incorrect inputs: arg_value is not int8")
			}
			inputs_interfaces = append(inputs_interfaces, int8(i))
		case "int16":
			i, err := strconv.ParseInt(arg_value, 10, 16)
			if err != nil {
				return nil, errors.New("incorrect inputs: arg_value is not int16")
			}
			inputs_interfaces = append(inputs_interfaces, int16(i))
		case "int32":
			i, err := strconv.ParseInt(arg_value, 10, 32)
			if err != nil {
				return nil, errors.New("incorrect inputs: arg_value is not int32")
			}
			inputs_interfaces = append(inputs_interfaces, int32(i))
		case "int64":
			i, err := strconv.ParseInt(arg_value, 10, 64)
			if err != nil {
				return nil, errors.New("incorrect inputs: arg_value is not int64")
			}
			inputs_interfaces = append(inputs_interfaces, int64(i))
		case "uint8":
			i, err := strconv.ParseInt(arg_value, 10, 8)
			if err != nil {
				return nil, errors.New("incorrect inputs: arg_value is not uint8")
			}
			inputs_interfaces = append(inputs_interfaces, big.NewInt(i))
		case "uint16":
			i, err := strconv.ParseInt(arg_value, 10, 16)
			if err != nil {
				return nil, errors.New("incorrect inputs: arg_value is not uint16")
			}
			inputs_interfaces = append(inputs_interfaces, big.NewInt(i))
		case "uint32":
			i, err := strconv.ParseInt(arg_value, 10, 32)
			if err != nil {
				return nil, errors.New("incorrect inputs: arg_value is not uint32")
			}
			inputs_interfaces = append(inputs_interfaces, big.NewInt(i))
		case "uint64":
			i, err := strconv.ParseInt(arg_value, 10, 64)
			if err != nil {
				return nil, errors.New("incorrect inputs: arg_value is not uint64")
			}
			inputs_interfaces = append(inputs_interfaces, big.NewInt(i))
		case "*big.Int":
			bi := new(big.Int)
			bi, _ = bi.SetString(arg_value, 10)
			if bi == nil {
				return nil, errors.New("incorrect inputs: " + arg_value + " not " + inputs_args[i].Type.String())
			}
			inputs_interfaces = append(inputs_interfaces, bi)
		case "[]*big.Int":
			var result []*big.Int

			result_array := strings.Split(arg_value, ",")

			for _, big_value := range result_array {
				bi := new(big.Int)
				bi, _ = bi.SetString(big_value, 10)
				if bi == nil {
					return nil, errors.New("incorrect inputs: " + arg_value + " not " + inputs_args[i].Type.String())
				}
				result = append(result, bi)
			}
			inputs_interfaces = append(inputs_interfaces, result)
		}
	}

	return inputs_interfaces, nil
}

func (w *EthWorker) ParseOutput(outputs []interface{}) (string, error) {

	if len(Containers.Containers[w.Container].Contracts[w.Contract].Abi.Methods[w.Endpoint].Outputs) == 0 {
		return "", nil
	}

	if Containers.Containers[w.Container] == nil || Containers.Containers[w.Container].Contracts[w.Contract] == nil {
		return "", errors.New("input values incorrect")
	}

	if len(Containers.Containers[w.Container].Contracts[w.Contract].Abi.Methods[w.Endpoint].Outputs) != 0 && Containers.Containers[w.Container].Contracts[w.Contract].OutputsInterfaces[w.Endpoint] == nil {
		return "", errors.New("input values incorrect")
	}

	output_args := Containers.Containers[w.Container].Contracts[w.Contract].Abi.Methods[w.Endpoint].Outputs

	if len(outputs) != len(output_args) {
		return "", errors.New("incorrect inputs")
	}

	var item_array []string

	for i := 0; i < len(outputs); i++ {

		switch output_args[i].Type.Type.String() {
		case "bool":
			item := strconv.FormatBool(*outputs[i].(*bool))

			item_array = append(item_array, item)

		case "[]bool":
			boolArray := *outputs[i].(*[]bool)
			var boolItems []string

			for _, bool_value := range boolArray {
				item := strconv.FormatBool(bool_value)
				boolItems = append(boolItems, item)
			}
			item := "[ " + strings.Join(boolItems, ",") + " ]"
			item_array = append(item_array, item)

		case "string":
			item_array = append(item_array, *outputs[i].(*string))
		case "[]string":
			array := *outputs[i].(*[]string)
			var items []string

			for _, value := range array {
				items = append(items, value)
			}
			item := "[ " + strings.Join(items, ",") + " ]"
			item_array = append(item_array, item)
		case "[]byte":
			array := *outputs[i].(*[]byte)
			var items []string

			for _, value := range array {
				items = append(items, string(value))
			}
			item := "[ " + strings.Join(items, ",") + " ]"
			item_array = append(item_array, item)
		case "[][]byte":
			array := *outputs[i].(*[][]byte)
			var items string

			for _, array2 := range array {

				var items2 []string

				for _, value := range array2 {
					items2 = append(items2, string(value))
				}
				item2 := "[ " + strings.Join(items2, ",") + " ]"
				items = items + "," + item2
			}
			item_array = append(item_array, items)
		case "common.Address":
			item := *outputs[i].(*common.Address)

			item_array = append(item_array, item.String())

		case "[]common.Address":
			addrArray := *outputs[i].(*[]common.Address)
			var addrItems []string

			for _, value := range addrArray {
				addrItems = append(addrItems, value.String())
			}
			item := "[ " + strings.Join(addrItems, ",") + " ]"
			item_array = append(item_array, item)
		case "common.Hash":
			item := *outputs[i].(*common.Hash)

			item_array = append(item_array, item.String())
		case "[]common.Hash":
			hashArray := *outputs[i].(*[]common.Hash)
			var hashItems []string

			for _, value := range hashArray {
				hashItems = append(hashItems, value.String())
			}
			item := "[ " + strings.Join(hashItems, ",") + " ]"
			item_array = append(item_array, item)
		case "int8":
			item := *outputs[i].(*int8)
			str := strconv.FormatInt(int64(item), 10)
			item_array = append(item_array, str)
		case "int16":
			item := *outputs[i].(*int16)
			str := strconv.FormatInt(int64(item), 10)
			item_array = append(item_array, str)
		case "int32":
			item := *outputs[i].(*int32)
			str := strconv.FormatInt(int64(item), 10)
			item_array = append(item_array, str)
		case "int64":
			item := *outputs[i].(*int64)
			str := strconv.FormatInt(item, 10)
			item_array = append(item_array, str)
		case "uint8":
			item := *outputs[i].(*uint8)
			str := strconv.FormatInt(int64(item), 10)
			item_array = append(item_array, str)
		case "uint16":
			item := *outputs[i].(*uint16)
			str := strconv.FormatInt(int64(item), 10)
			item_array = append(item_array, str)
		case "uint32":
			item := *outputs[i].(*uint32)
			str := strconv.FormatInt(int64(item), 10)
			item_array = append(item_array, str)
		case "uint64":
			item := *outputs[i].(*uint64)
			str := strconv.FormatInt(int64(item), 10)
			item_array = append(item_array, str)
		case "*big.Int":
			item := *outputs[i].(**big.Int)
			item_array = append(item_array, item.String())
		case "[]*big.Int":
			bigArray := *outputs[i].(*[]*big.Int)
			var items []string
			for _, v := range bigArray {
				items = append(items, v.String())
			}
			item := "[ " + strings.Join(items, ",") + " ]"
			item_array = append(item_array, item)
		}
	}
	return strings.Join(item_array, " , "), nil
}

func Bind(dirname, solcfile string) (*ContractContainers, error) {
	result := &ContractContainers{
		Containers: make(map[string]*ContractContainer),
	}

	allfiles, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, errors.Wrap(err, "error ioutil.ReadDir")
	}

	for _, v := range allfiles {
		if v.IsDir() {
			continue
		}
		if hasSuffixCaseInsensitive(v.Name(), ".sol") {
			contracts, err := compiler.CompileSolidity(solcfile, dirname+string(os.PathSeparator)+v.Name())
			if err != nil {
				return nil, errors.Wrap(err, "CompileSolidity")
			}

			c := &ContractContainer{
				ContainerName: v.Name(),
				Contracts:     make(map[string]*Contract),
			}

			for name, contract := range contracts {
				a, _ := json.Marshal(contract.Info.AbiDefinition)
				ab, err := abi.JSON(strings.NewReader(string(a)))
				if err != nil {
					return nil, errors.Wrap(err, "abi.JSON")
				}
				nameParts := strings.Split(name, ":")

				var ab_keys []string

				ouputs_map := make(map[string][]interface{})
				inputs_map := make(map[string][]interface{})

				for key, method := range ab.Methods {
					ab_keys = append(ab_keys, key)

					var o []interface{}
					var i []interface{}

					for _, v := range method.Outputs {
						var ar interface{}

						switch v.Type.Type.String() {
						case "bool":
							ar = new(bool)
						case "[]bool":
							ar = new([]bool)
						case "string":
							ar = new(string)
						case "[]string":
							ar = new([]string)
						case "[]byte":
							ar = new([]byte)
						case "[][]byte":
							ar = new([][]byte)
						case "common.Address":
							ar = new(common.Address)
						case "[]common.Address":
							ar = new([]common.Address)
						case "common.Hash":
							ar = new(common.Hash)
						case "[]common.Hash":
							ar = new([]common.Hash)
						case "int8":
							ar = new(int8)
						case "int16":
							ar = new(int16)
						case "int32":
							ar = new(int32)
						case "int64":
							ar = new(int64)
						case "uint8":
							ar = new(uint8)
						case "uint16":
							ar = new(uint16)
						case "uint32":
							ar = new(uint32)
						case "uint64":
							ar = new(uint64)
						case "*big.Int":
							ar = new(*big.Int)
						case "[]*big.Int":
							ar = new([]*big.Int)
						default:
							return nil, errors.Errorf("unsupported type: %s", v.Type.Type.String())
						}

						o = append(o, ar)

					}

					ouputs_map[method.Name] = o

					for _, v := range method.Inputs {
						var ar interface{}

						switch v.Type.Type.String() {
						case "bool":
							ar = new(bool)
						case "[]bool":
							ar = new([]bool)
						case "string":
							ar = new(string)
						case "[]string":
							ar = new([]string)
						case "[]byte":
							ar = new([]byte)
						case "[][]byte":
							ar = new([][]byte)
						case "common.Address":
							ar = new(common.Address)
						case "[]common.Address":
							ar = new([]common.Address)
						case "common.Hash":
							ar = new(common.Hash)
						case "[]common.Hash":
							ar = new([]common.Hash)
						case "int8":
							ar = new(int8)
						case "int16":
							ar = new(int16)
						case "int32":
							ar = new(int32)
						case "int64":
							ar = new(int64)
						case "uint8":
							ar = new(uint8)
						case "uint16":
							ar = new(uint16)
						case "uint32":
							ar = new(uint32)
						case "uint64":
							ar = new(uint64)
						case "*big.Int":
							ar = new(*big.Int)
						case "[]*big.Int":
							ar = new([]*big.Int)
						default:
							return nil, errors.Errorf("unsupported type: %s", v.Type.Type.String())
						}
						i = append(i, ar)
					}

					inputs_map[method.Name] = i
				}
				sort.Strings(ab_keys)

				con := &Contract{
					Name:              nameParts[len(nameParts)-1],
					Abi:               ab,
					AbiJson:           string(a),
					Bin:               contract.Code,
					SortKeys:          ab_keys,
					OutputsInterfaces: ouputs_map,
					InputsInterfaces:  inputs_map,
				}

				c.ContractNames = append(c.ContractNames, nameParts[len(nameParts)-1])
				c.Contracts[nameParts[len(nameParts)-1]] = con
			}
			sort.Strings(c.ContractNames)

			result.ContainerNames = append(result.ContainerNames, c.ContainerName)
			result.Containers[c.ContainerName] = c
		}
	}

	sort.Strings(result.ContainerNames)

	return result, err

}

func hasSuffixCaseInsensitive(s, suffix string) bool {
	return len(s) >= len(suffix) && strings.ToLower(s[len(s)-len(suffix):]) == suffix
}
