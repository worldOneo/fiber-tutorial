<script>
	import Input from "./components/input.svelte";
	let name, password, regpassword, reppassword;
	let login = true;
	let loggedin = false;
	let username = "Nicht eingeloggt!";

	let response = "Loading...";
	fetch("http://localhost:5000/api/v1/time", {
		headers: {
			accept:
				"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
			"accept-language": "en-US,en;q=0.9",
			"cache-control": "max-age=0",
			"sec-fetch-dest": "document",
			"sec-fetch-mode": "navigate",
			"sec-fetch-site": "none",
			"sec-fetch-user": "?1",
			"sec-gpc": "1",
			"upgrade-insecure-requests": "1",
		},
		referrerPolicy: "strict-origin-when-cross-origin",
		body: null,
		method: "GET",
		mode: "cors",
		credentials: "omit",
	}).then(r => {
		r.text().then(v => {
			response = v;
		})
	});
</script>

<main>
	<h1>API {response}</h1>
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
			<button>Login</button>
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
			<button>Login</button>
			<button on:click={(_) => (login = !login)}>Zum Login</button>
		</div>
	{:else}
		<h1>Willkommen, {username}</h1>
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
