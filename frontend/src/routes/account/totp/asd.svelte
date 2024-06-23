<script lang="ts">
	import { registerTOTPBegin, registerTOTPFinish, registerTOTPGetSecret } from '$lib/client';
	import { onMount } from 'svelte';

	let promise = new Promise(() => {});
	let totpSecret = '';
	let recoveryCodes: string[] = [];
	let error = '';

	onMount(() => {
		promise = (async () => {
			await registerTOTPBegin().catch((error) => {
				error = error;
			});
			await registerTOTPGetSecret()
				.then((response) => {
					totpSecret = response.totp_secret;
				})
				.catch((error) => {
					error = error;
				});
		})();
	});

	const onSubmit = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);
		const code = formData.get('code') as string;
		registerTOTPFinish(code)
			.then((response) => {
				recoveryCodes = response.recovery_codes;
			})
			.catch((error) => {
				error = error;
			});
	};
</script>

{#await promise}
	<p>loading...</p>
{:then}
	{#if error}
		<p style="color: red">{error}</p>
	{/if}

	<p>TOTP Secret: {totpSecret}</p>
	<img src="/api/auth/2fa/totp/qr" alt="QR Code" />
	<form action="/api/auth/2fa/totp/confirm" on:submit|preventDefault={onSubmit}>
		<label for="code">Code</label>
		<input type="text" name="code" id="code" />
		<button type="submit">Submit</button>
	</form>

	{#if recoveryCodes.length > 0}
		<h2>Recovery Codes</h2>
		<ul>
			{#each recoveryCodes as code}
				<li>{code}</li>
			{/each}
		</ul>
	{/if}
{:catch error}
	<p style="color: red">{error.message}</p>
{/await}
