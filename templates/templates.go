package templates

const (
	PageTemplateHeader = `<html>
<head>
	<title>Contract GUI</title>
	<meta charset="utf8" />
	<link href="http://fonts.googleapis.com/css?family=Open+Sans" rel="stylesheet" type="text/css">
	<style>
		body{font-size: 9pt; font-family: 'Open Sans', sans-serif;}
		a{text-decoration: none}
		a:hover {text-decoration: underline}
		.keys > span:hover { background: #f0f0f0; }
		span:target { background: #ccffcc; }

		#funcs {
			font-family: "Trebuchet MS", Arial, Helvetica, sans-serif;
			border-collapse: collapse;
			width: 100%;
		}

		#funcs td, #funcs th {
			border: 1px solid #ddd;
			padding: 8px;
		}

		#funcs tr:nth-child(even){background-color: #f2f2f2;}

		#funcs tr:hover {background-color: #ddd;}

		#funcs th {
			padding-top: 12px;
			padding-bottom: 12px;
			text-align: left;
			background-color: #4CAF50;
			color: white;
		}

		#ether {
			font-family: "Trebuchet MS", Arial, Helvetica, sans-serif;
			border-collapse: collapse;
			width: 100%;
		}

		#ether td, #funcs th {
			border: 1px solid #ddd;
			padding: 8px;
		}

		#ether tr:nth-child(even){background-color: #f2f2f2;}

		#ether tr:hover {background-color: #ddd;}

		#ether th {
			padding-top: 12px;
			padding-bottom: 12px;
			text-align: left;
			background-color: #4CAF50;
			color: white;
		}

		#header {
			padding-top: 12px;
			padding-bottom: 12px;
		}

		.container{overflow:hidden;width:100%}
		.box{white-space:nowrap}
		.box div{display:inline-block; float:left; overflow-y: auto; height: 400px; width: 50%;}

		#log_area {
			margin: 0px; width: 100%; height: 160px; resize: none;
		}

		#div-ether {
			float:left; overflow-y: auto; width: 50%;
		}

		.brd {
			border: 4px double black;
			padding: 10px;
		}

	</style>
</head>
<body>
`

	PageTemplateFutter = `
</body>
</html>
`

	HeaderContainer = `
<div>
	<div>
		<div>
			<a href="/login">login</a>
		</div>
		<div>
			<a href="/upload">upload</a>
		</div>
		{{if .}}
		<div>
			<table id="header">
			<tr>
				<td>you address:</td>
				<td>{{.Address}}</td>
			</tr>
			<tr>
				<td>balance:</td>
				<td>{{.EthBalance}}</td>
			</tr>
			<tr>
				<td>file:</td>
				<td>{{.Container}}</td>
			</tr>
			<tr>
				<td>contract:</td>
				<td>{{.Contract}}</td>
			</tr>
			<tr>
				<td>contract address:</td>
				<td>{{.ContractAddress}}</td>
			</tr>
			</table>
		</div>
		{{end}}
	</div>
</div>
`

	SelectContainer = `
<div class="brd">
	<form action="/upload" method="post">
		<p>Select Container</p>
		<p>
			<select name="container">
				{{range .}}
					<option value="{{.}}">{{.}}</option>
				{{end}}
	   		</select>
		</p>
	   <p><input type="submit" value="Send"></p>
  </form>
</div>
`

	SelectContract = `
<div class="brd">
	<form action="/upload" method="post">
		<p>Select Contract</p>
		<p>
			<select name="contract">
				{{range .}}
					<option value="{{.}}">{{.}}</option>
				{{end}}
	   		</select>
			<p><input type="checkbox" name="deploy"> Deploy</p>
		</p>
		<p><input type="text" name=address title="contract address" placeholder="contract address"></p>
	    <p><input type="submit" value="Send"></p>
  </form>
</div>
`

	DeployTemplate = `
<div>
<form action="/deploy" method="post">
	{{if .Inputs}}
			{{with .Inputs}}
			{{range $i, $v := .}}
			<p>
			<input type="text" name={{$i}} title="{{$v.Name}} {{$v.Type}}" placeholder="{{$v.Name}} {{$v.Type}}">
			</p>
			{{end}}
			{{end}}
		{{end}}

	<p><input type="submit" value="deploy contract" title="deploy contract"></p>
</form>
</div>
`

	FormStart = `
<div class="container">
	<div>
		<textarea id="log_area" rows="10" cols="45" name="log" disabled>
{{.}}
		</textarea>
	</div>
	<div class="box">

		<div>
			<table id="funcs">
			<tbody>
		`
	FormFinish = `
			</tbody>
			</table>
		</div>

		<div id="div-ether">
			<table id="ether">
			<tbody>
					<tr>
						<form action="/eth?endpoint=balance" method="post">
							<td>Balance</td>
							<td><input type="text" name="1" title="addr address" placeholder="addr address"></td>
							<td><input type="submit" value="balance"></td>
						</form>
					</tr>
					<tr>
						<form action="/eth?endpoint=gas_price" method="post">
							<td>Gas price</td>
							<td></td>
							<td><input type="submit" value="gas_price"></td>
						</form>
					</tr>
					<tr>
						<form action="/eth?endpoint=last_block" method="post">
							<td>Last block</td>
							<td></td>
							<td><input type="submit" value="last_block"></td>
						</form>
					</tr>
					<tr>
						<form action="/eth?endpoint=gas_limit" method="post">
							<td>Ethereum network gas limit</td>
							<td></td>
							<td><input type="submit" value="gas_limit"></td>
						</form>
					</tr>
					<tr>
						<form action="/eth?endpoint=time" method="post">
							<td>Ethereum network time</td>
							<td></td>
							<td><input type="submit" value="time"></td>
						</form>
					</tr>
					<tr>
						<form action="/eth?endpoint=difficulty" method="post">
							<td>Ethereum network difficulty</td>
							<td></td>
							<td><input type="submit" value="difficulty"></td>
						</form>
					</tr>
					<tr>
						<form action="/eth?endpoint=transfer" method="post">
							<td>Transfer</td>
							<td>
								<input type="text" name="1" title="to address" placeholder="to address">
								<input type="text" name="2" title="value uint256" placeholder="value uint256">
							</td>
							<td><input type="submit" value="transfer"></td>
						</form>
					</tr>
					<tr>
						<form action="/eth?endpoint=adjusttime" method="post">
							<td>Adjust time</td>
							<td>
								<input type="text" name="1" title="increase the time" placeholder="increase the time">
							</td>
							<td><input type="submit" value="adjust time"></td>
						</form>
					</tr>
			</tbody>
			</table>
		</div>
	</div>
</div>
`

	MethodTemplate = `
<tr>
<form action="/{{if .Const}}public{{else}}private{{end}}?endpoint={{.Name}}" method="post">
	<td>{{.Sig}}:</td>

	<td>{{if .Inputs}}
			{{with .Inputs}}
			{{range $i, $v := .}}
			<input type="text" name={{$i}} title="{{$v.Name}} {{$v.Type}}" placeholder="{{$v.Name}} {{$v.Type}}">
			{{end}}
			{{end}}
		{{end}}
	</td>
	<td><input type="submit" value={{.Name}} title="{{.String}}"></td>
</form>
</tr>
`

	LoginTemplate = `
<div class="main-login">
	<form action="/update" method="post">
		<div class="field">
			<label for="pk">Private Key</label>
			<input type="text" name="login-pk" id="pk">
		</div>

		<div class="field">
			<input type="submit" value=login title="ok">
		</div>
	</form>
</div>
`

//	UploadTemplate = `
//<div class="load">
//	<form action="/upload" enctype="multipart/form-data"  method="post" >
//		<div class="field">
//			<label for="sol-file">Sol file</label>
//			<input type="file" name="sol-file">
//		</div>
//
//		<div class="field">
//			<input type="submit" value="upload" title="upload">
//		</div>
//	</form>
//</div>
//`
//MethodTemplate     = `<tr><td>{{.Name}}:<input type="text" name="{{.Name}}"></td><td><input type="submit" value={{.Name}}></td>INPUTS:{{with .Inputs}}{{range .}}{{.}}{{end}}{{end}} OUTPUTS:{{with .Outputs}}{{range .}}{{.}}{{end}}{{end}}</tr>`
//MethodFormTemplate = `METHOD name: {{.Name}} INPUTS:{{with .Inputs}}{{range .}}{{.}}{{end}}{{end}} OUTPUTS:{{with .Outputs}}{{range .}}{{.}}{{end}}{{end}}{{"\n"}}`
)

var (
	WriteResult = `Nonce %d:
From: %s
To: %s
Value: %s
Gas price: %s
Gas Used: %s
Cost/Fee: %s
Status: %d
Transaction Hash: %s`

	DeployResult = `Nonce %d:
From: %s
Contract Address: %s
Gas price: %s
Gas Used: %s
Cost/Fee: %s
Status: %d
Transaction Hash: %s`
)
