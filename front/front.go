package front

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"html/template"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"ethereum-front/abi/bind"
	"ethereum-front/ether"
	"ethereum-front/templates"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"net/url"
	"strconv"
)

func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	//dummy
}

func SetCookieHandler(w http.ResponseWriter, r *http.Request) {

	f := func(name, value string, duration int64) *http.Cookie {
		cookie := &http.Cookie{}
		cookie.Name = name
		cookie.Value = value

		if duration > 0 {
			cookie.Expires = time.Now().Add(time.Second * time.Duration(duration))
		}
		return cookie
	}
	if r.Method == "POST" {
		r.ParseForm()
		private_key := r.Form.Get("login-pk")
		if private_key != "" {
			cookie := f("private_key", private_key, 0)
			http.SetCookie(w, cookie)
		}
		container := r.Form.Get("container")
		if container != "" {
			cookie := f("container", container, 0)
			http.SetCookie(w, cookie)
		}
		contract := r.Form.Get("contract")
		if contract != "" {
			cookie := f("contract", contract, 0)
			http.SetCookie(w, cookie)
		}
		address := r.Form.Get("address")
		if address != "" {
			cookie := f("address", address, 0)
			http.SetCookie(w, cookie)
		}
		r.Method = "GET"
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func MainPage(w http.ResponseWriter, r *http.Request) {

	key, err := r.Cookie("private_key")
	if err != nil || key == nil || key.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	container, err := r.Cookie("container")
	if err != nil || container == nil || container.Value == "" || ether.Containers.Containers[container.Value] == nil {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	contract, err := r.Cookie("contract")
	if err != nil || contract == nil || contract.Value == "" || ether.Containers.Containers[container.Value].Contracts[contract.Value] == nil {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	address, err := r.Cookie("address")
	if err != nil || address == nil || address.Value == "" {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	informer := ether.NewEthWorker(
		container.Value,
		contract.Value,
		"",
		key.Value,
		address.Value,
		url.Values{},
	)
	info, err := informer.Info()
	if err != nil {
		log.Printf("error info %s", err.Error())
	}

	fmt.Fprint(w, templates.PageTemplateHeader)

	tInfo := template.New("info")
	tInfo.Parse(templates.HeaderContainer)
	tInfo.Execute(w, info)

	t2 := template.New("Textarea")
	t2.Parse(templates.FormStart)
	t2.Execute(w, "")

	t := template.New("Methods")
	t, _ = t.Parse(templates.MethodTemplate)

	for _, v := range ether.Containers.Containers[container.Value].Contracts[contract.Value].SortKeys {
		t.Execute(w, ether.Containers.Containers[container.Value].Contracts[contract.Value].Abi.Methods[v])
	}
	t3 := template.New("body2")
	t3.Parse(templates.FormFinish)
	t3.Execute(w, "table2")
	fmt.Fprint(w, templates.PageTemplateFutter)
}

func Private(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	endpoint := r.Form.Get("endpoint")
	if endpoint == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	key, err := r.Cookie("private_key")
	if err != nil || key == nil || key.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	container, err := r.Cookie("container")
	if err != nil || container == nil || container.Value == "" || ether.Containers.Containers[container.Value] == nil {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	contract, err := r.Cookie("contract")
	if err != nil || contract == nil || contract.Value == "" || ether.Containers.Containers[container.Value].Contracts[contract.Value] == nil {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	address, err := r.Cookie("address")
	if err != nil || address == nil || address.Value == "" {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	writer := ether.NewEthWorker(
		container.Value,
		contract.Value,
		endpoint,
		key.Value,
		address.Value,
		r.Form,
	)

	var responce string

	result, err := writer.Transact()
	if err != nil {
		responce = fmt.Sprintf("Error: %s", err.Error())
	} else {
		responce = fmt.Sprintf("Result: %s", result)
	}

	info, err := writer.Info()

	fmt.Fprint(w, templates.PageTemplateHeader)

	tInfo := template.New("info")
	tInfo.Parse(templates.HeaderContainer)
	tInfo.Execute(w, info)

	t2 := template.New("Textarea")
	t2.Parse(templates.FormStart)
	t2.Execute(w, endpoint+" : "+responce)

	t := template.New("Methods")
	t, _ = t.Parse(templates.MethodTemplate)

	for _, v := range ether.Containers.Containers[container.Value].Contracts[contract.Value].SortKeys {
		t.Execute(w, ether.Containers.Containers[container.Value].Contracts[contract.Value].Abi.Methods[v])
	}

	t3 := template.New("body2")
	t3.Parse(templates.FormFinish)
	t3.Execute(w, "table2")

	fmt.Fprint(w, templates.PageTemplateFutter)

}

func Public(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	endpoint := r.Form.Get("endpoint")
	if endpoint == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	key, err := r.Cookie("private_key")
	if err != nil || key == nil || key.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	container, err := r.Cookie("container")
	if err != nil || container == nil || container.Value == "" || ether.Containers.Containers[container.Value] == nil {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	contract, err := r.Cookie("contract")
	if err != nil || contract == nil || contract.Value == "" || ether.Containers.Containers[container.Value].Contracts[contract.Value] == nil {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	address, err := r.Cookie("address")
	if err != nil || address == nil || address.Value == "" {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	reader := ether.NewEthWorker(
		container.Value,
		contract.Value,
		endpoint,
		key.Value,
		address.Value,
		r.Form,
	)

	var responce string

	result, err := reader.Call()
	if err != nil {
		responce = fmt.Sprintf("Error: %s", err.Error())
	} else {
		responce = fmt.Sprintf("Result: %s", result)
	}

	info, err := reader.Info()

	fmt.Fprint(w, templates.PageTemplateHeader)

	tInfo := template.New("info")
	tInfo.Parse(templates.HeaderContainer)
	tInfo.Execute(w, info)

	t2 := template.New("Textarea")
	t2.Parse(templates.FormStart)
	t2.Execute(w, endpoint+" : "+responce)

	t := template.New("Methods")
	t, _ = t.Parse(templates.MethodTemplate)

	for _, v := range ether.Containers.Containers[container.Value].Contracts[contract.Value].SortKeys {
		t.Execute(w, ether.Containers.Containers[container.Value].Contracts[contract.Value].Abi.Methods[v])
	}

	t3 := template.New("body2")
	t3.Parse(templates.FormFinish)
	t3.Execute(w, "table2")

	fmt.Fprint(w, templates.PageTemplateFutter)
}

func EthPage(w http.ResponseWriter, r *http.Request) {
	var result string

	r.ParseForm()
	endpoint := r.Form.Get("endpoint")

	c, err := r.Cookie("private_key")
	if err != nil || c == nil || c.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pk := strings.TrimPrefix(c.Value, "0x")
	key, err := crypto.HexToECDSA(pk)
	if err != nil {
		log.Println(err.Error())
		return
	}
	auth := bind.NewKeyedTransactor(key)

	switch endpoint {
	case "balance":
		addr := r.Form.Get("1")
		if !common.IsHexAddress(addr) {
			result = addr + " : is not address"
			break
		}

		switch v := ether.Client.(type) {
		case *ethclient.Client:
			balance, err := v.BalanceAt(
				context.Background(),
				common.HexToAddress(addr),
				nil,
			)
			if err != nil {
				result = "error: " + err.Error()
				break
			}
			result = "address: " + addr + " , balance: " + balance.String()
		case *backends.SimulatedBackend:
			balance, err := v.BalanceAt(
				context.Background(),
				common.HexToAddress(addr),
				nil,
			)
			if err != nil {
				result = "error: " + err.Error()
				break
			}
			result = "address: " + addr + " , balance: " + balance.String()
		}

	case "gas_price":
		gasprice, err := ether.Client.SuggestGasPrice(context.Background())
		if err != nil {
			result = "error: " + err.Error()
			break
		}
		result = "Gas price: " + gasprice.String()
	case "last_block":
		switch v := ether.Client.(type) {
		case *ethclient.Client:
			block, err := v.BlockByNumber(context.Background(), nil)
			if err != nil {
				result = "error: " + err.Error()
				break
			}
			result = "Last block number: " + block.Number().String()
		case *backends.SimulatedBackend:
			result = "It does not work on an emulator"
		}

	case "gas_limit":
		switch v := ether.Client.(type) {
		case *ethclient.Client:
			block, err := v.BlockByNumber(context.Background(), nil)
			if err != nil {
				result = "error: " + err.Error()
				break
			}
			result = "Gas limit: " + block.GasLimit().String()
		case *backends.SimulatedBackend:
			result = "It does not work on an emulator"
		}

	case "time":

		switch v := ether.Client.(type) {
		case *ethclient.Client:
			block, err := v.BlockByNumber(context.Background(), nil)
			if err != nil {
				result = "error: " + err.Error()
				break
			}
			result = "Time: " + block.Time().String()
		case *backends.SimulatedBackend:
			result = "It does not work on an emulator"
		}

	case "difficulty":
		switch v := ether.Client.(type) {
		case *ethclient.Client:
			block, err := v.BlockByNumber(context.Background(), nil)
			if err != nil {
				result = "error: " + err.Error()
				break
			}
			result = "Difficulty: " + block.Difficulty().String()
		case *backends.SimulatedBackend:
			result = "It does not work on an emulator"
		}

	case "adjusttime":
		switch v := ether.Client.(type) {
		case *ethclient.Client:
			result = "It does not work on real connection"
		case *backends.SimulatedBackend:
			int_string := r.Form.Get("1")
			i, err := strconv.ParseInt(int_string, 10, 64)
			if err != nil {
				result = "error: " + err.Error()
				break
			}

			if i < 0 {
				result = fmt.Sprintf("error: %d < 0", i)
				break
			}
			if err := v.AdjustTime(time.Duration(i) * time.Second); err != nil {
				result = "error: " + err.Error()
				break
			}
			v.Commit()
			result = "adjustment complete"
		}

	case "transfer":
		to_addr := r.Form.Get("1")
		if !common.IsHexAddress(to_addr) {
			result = to_addr + " : is not address"
			break
		}
		value := r.Form.Get("2")

		bigValue := new(big.Int)
		bigValue, _ = bigValue.SetString(value, 10)

		bigGaslimit := new(big.Int)
		bigGaslimit, _ = bigGaslimit.SetString(ether.GasLimit.String(), 10)

		gasprice, err := ether.Client.SuggestGasPrice(context.Background())
		if err != nil {
			result = "error: " + err.Error()
		}

		bigGasprice := new(big.Int)
		bigGasprice, _ = bigGasprice.SetString(gasprice.String(), 10)

		nonce, err := ether.Client.PendingNonceAt(context.Background(), auth.From)

		rawTx := types.NewTransaction(uint64(nonce), common.HexToAddress(to_addr), bigValue, bigGaslimit, bigGasprice, nil)

		signedTx, err := auth.Signer(types.HomesteadSigner{}, auth.From, rawTx)
		if err != nil {
			result = "error: " + err.Error()
		}

		if err := ether.Client.SendTransaction(context.Background(), signedTx); err != nil {
			result = "error: " + err.Error()
		}

		var receipt *types.Receipt

		switch v := ether.Client.(type) {
		case *backends.SimulatedBackend:
			v.Commit()
			receipt, err = v.TransactionReceipt(context.Background(), signedTx.Hash())
			if err != nil {
				result = errors.Wrap(err, "transaction receipt").Error()
			}

		case *ethclient.Client:
			receipt, err = bind.WaitMined(context.Background(), v, signedTx)
			if err != nil {
				result = errors.Wrap(err, "transaction receipt").Error()
			}
		}

		if err == nil {
			result = fmt.Sprintf(templates.WriteResult,
				signedTx.Nonce(),
				auth.From.String(),
				signedTx.To().String(),
				signedTx.Value().String(),
				signedTx.GasPrice().String(),
				receipt.GasUsed.String(),
				new(big.Int).Mul(receipt.GasUsed, signedTx.GasPrice()),
				receipt.Status,
				receipt.TxHash.String(),
			)
		}
	}

	container, err := r.Cookie("container")
	if err != nil || container == nil || container.Value == "" || ether.Containers.Containers[container.Value] == nil {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	contract, err := r.Cookie("contract")
	if err != nil || contract == nil || contract.Value == "" || ether.Containers.Containers[container.Value].Contracts[contract.Value] == nil {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	address, err := r.Cookie("address")
	if err != nil {
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}

	var address_current string

	if address != nil {
		address_current = address.Value
	}

	informer := ether.NewEthWorker(
		container.Value,
		contract.Value,
		"",
		c.Value,
		address_current,
		url.Values{},
	)

	info, err := informer.Info()
	if err != nil {
		log.Printf("error info %s", err.Error())
	}

	fmt.Fprint(w, templates.PageTemplateHeader)

	tInfo := template.New("info")
	tInfo.Parse(templates.HeaderContainer)
	tInfo.Execute(w, info)

	t2 := template.New("Textarea")
	t2.Parse(templates.FormStart)
	t2.Execute(w, result)

	t := template.New("Methods")
	t, _ = t.Parse(templates.MethodTemplate)

	for _, v := range ether.Containers.Containers[container.Value].Contracts[contract.Value].SortKeys {
		t.Execute(w, ether.Containers.Containers[container.Value].Contracts[contract.Value].Abi.Methods[v])
	}

	t3 := template.New("body2")
	t3.Parse(templates.FormFinish)
	t3.Execute(w, "table2")
	fmt.Fprint(w, templates.PageTemplateFutter)
}

func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, templates.PageTemplateHeader)
	fmt.Fprint(w, templates.LoginTemplate)
	fmt.Fprint(w, templates.PageTemplateFutter)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	var container string

	if r.Method == "POST" {
		r.ParseForm()
		contract := r.Form.Get("contract")
		container = r.Form.Get("container")
		address := r.Form.Get("address")
		deploy := r.Form.Get("deploy")

		if container != "" && ether.Containers.Containers[container] != nil {
			cookie := &http.Cookie{Name: "container", Value: container}
			http.SetCookie(w, cookie)
		}
		if contract != "" {

			c, err := r.Cookie("container")
			if err != nil || c == nil || c.Value == "" {
				http.Redirect(w, r, "/upload", http.StatusSeeOther)
				return
			}
			if ether.Containers.Containers[c.Value].Contracts[contract] != nil {
				cookie := &http.Cookie{Name: "contract", Value: contract}
				http.SetCookie(w, cookie)
			}
		}
		if deploy == "on" {
			r.Method = "GET"
			http.Redirect(w, r, "/deploy", http.StatusSeeOther)
			return
		}
		if address != "" {
			c1, err := r.Cookie("container")
			if err != nil || c1 == nil || c1.Value == "" {
				http.Redirect(w, r, "/upload", http.StatusSeeOther)
				return
			}
			c2, err := r.Cookie("contract")
			if err != nil || c2 == nil || c2.Value == "" {
				http.Redirect(w, r, "/upload", http.StatusSeeOther)
				return
			}
			cookie := &http.Cookie{Name: "address", Value: address}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	fmt.Fprint(w, templates.PageTemplateHeader)
	t0 := template.New("container")
	t0.Parse(templates.SelectContainer)
	t0.Execute(w, ether.Containers.ContainerNames)

	if container != "" && ether.Containers.Containers[container] != nil {
		t1 := template.New("contract")
		t1.Parse(templates.SelectContract)
		t1.Execute(w, ether.Containers.Containers[container].ContractNames)
	}
	fmt.Fprint(w, templates.PageTemplateFutter)
}

func Deploy(w http.ResponseWriter, r *http.Request) {
	c1, err := r.Cookie("container")
	if err != nil || c1 == nil || c1.Value == "" {
		r.Method = "GET"
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}
	c2, err := r.Cookie("contract")
	if err != nil || c2 == nil || c2.Value == "" {
		r.Method = "GET"
		http.Redirect(w, r, "/upload", http.StatusSeeOther)
		return
	}
	c3, err := r.Cookie("private_key")
	if err != nil || c3 == nil || c3.Value == "" {
		r.Method = "GET"
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		deployer := ether.NewEthWorker(c1.Value, c2.Value, "", c3.Value, "", r.Form)
		deployer.New = true

		var result string
		result, address, err := deployer.Deploy()
		if err != nil {
			result = "deploy error: " + err.Error()
		}
		if address != "" {
			cookie := &http.Cookie{Name: "address", Value: address}
			http.SetCookie(w, cookie)
		}

		info, err := deployer.Info()
		fmt.Fprint(w, templates.PageTemplateHeader)

		tInfo := template.New("info")
		tInfo.Parse(templates.HeaderContainer)
		tInfo.Execute(w, info)

		t2 := template.New("Textarea")
		t2.Parse(templates.FormStart)
		t2.Execute(w, result)

		t := template.New("Methods")
		t, _ = t.Parse(templates.MethodTemplate)

		for _, v := range ether.Containers.Containers[c1.Value].Contracts[c2.Value].SortKeys {
			t.Execute(w, ether.Containers.Containers[c1.Value].Contracts[c2.Value].Abi.Methods[v])
		}

		t3 := template.New("body2")
		t3.Parse(templates.FormFinish)
		t3.Execute(w, "table2")

		fmt.Fprint(w, templates.PageTemplateFutter)
		return
	}

	t := template.New("Constructor")
	t, _ = t.Parse(templates.DeployTemplate)
	t.Execute(w, ether.Containers.Containers[c1.Value].Contracts[c2.Value].Abi.Constructor)

}

func Start(connect_url, sol_path, keystore_path string, port int, gaslimit int64, solc string) {

	ether.GasLimit = big.NewInt(gaslimit)
	params.GenesisGasLimit = big.NewInt(gaslimit)
	params.TargetGasLimit = new(big.Int).Set(big.NewInt(gaslimit))

	if connect_url == "" {
		alloc := make(core.GenesisAlloc)

		b1 := new(big.Int)
		fmt.Sscan("1000000000000000000000000000000000000000000000000000", b1)

		ks := keystore.NewKeyStore(
			keystore_path,
			keystore.LightScryptN,
			keystore.LightScryptP,
		)

		all_acc := ks.Accounts()
		for _, v := range all_acc {
			jsonAcc, _ := ks.Export(v, "", "")
			auth, _ := bind.NewTransactor(strings.NewReader(string(jsonAcc)), "")
			alloc[auth.From] = core.GenesisAccount{Balance: b1}
		}
		ether.Client = backends.NewSimulatedBackend(alloc)

	} else {
		conn, err := ethclient.Dial(connect_url)
		if err != nil {
			panic(err.Error())
		}
		ether.Client = conn
	}

	c, err := ether.Bind(sol_path, solc)
	if err != nil {
		panic(err.Error())
	}
	ether.Containers = c

	http.HandleFunc("/", MainPage)
	http.HandleFunc("/eth", EthPage)
	http.HandleFunc("/private", Private)
	http.HandleFunc("/public", Public)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/upload", Upload)
	http.HandleFunc("/update", SetCookieHandler)
	http.HandleFunc("/deploy", Deploy)
	http.HandleFunc("/favicon.ico", FaviconHandler)
	log.Println("Listening test frontend")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
