<script>
	import Input from "./components/input.svelte";
	import APIRequest from "./api";
	let name, password, regpassword, reppassword, token;
	let login = true;
	let loggedin = false;
	let username = "Nicht eingeloggt!";

	let response = "Loading...";

	function registerUser() {
		APIRequest("createuser", "POST", {username: name, password: regpassword}, m => {
			if(m.success) {
				alert(`Successful: ${m.message}`);
			} else {
				alert(`Failed: ${m.message}`);
			}
		});
	}

	function loginUser() {
		APIRequest("generatetoken", "POST", {username: name, password: password}, m => {
			if(m.success) {
				token = m.message
				loggedin = true
				username = name
			} else {
				alert(m.message)
			}
		})
	}
</script>

<main>
	{#if login && !loggedin}
		<h1>Login</h1>
		<div>
			<Input style="width: 100%" lable="Name" bind:value={name} />
			<Input
				style="width: 100%"
				lable="Passwort"
				bind:value={password}
				type="password"
			/>
			<button on:click={loginUser}>Login</button>
			<button on:click={(_) => (login = !login)}>Zum Registrieren</button>
		</div>
	{:else if !loggedin}
		<h1>Register</h1>
		<div>
			<Input style="width: 100%" lable="Name" bind:value={name} />
			<Input
				style="width: 100%"
				lable="Passwort"
				bind:value={regpassword}
				type="password"
			/>
			<Input
				style="width: 100%"
				lable="Passwort wiederholen"
				bind:value={reppassword}
				type="password"
			/>
			<button disabled={regpassword != reppassword} on:click={registerUser}>Registrieren</button>
			<button on:click={(_) => (login = !login)}>Zum Login</button>
		</div>
	{:else}
		<h1>Willkommen, {username}</h1>
		<h2>Dein Token ist '{token}'</h2>
	{/if}
</main>

<style>
	main {
		text-align: center;
		padding: 1em;
		margin: 0 auto;
	}
	div {
		margin: 0 auto;
		display: block;
		width: 20em;
	}
	button {
		width: 100%;
	}
</style>
