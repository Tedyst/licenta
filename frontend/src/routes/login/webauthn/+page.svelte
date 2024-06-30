<script lang="ts">
	import type { PageData } from './$types';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		JSONtoPublicKeyCredentialRequestOptions,
		LoginPublicKeyCredentialToJSON
	} from '$lib/webauthn';
	import { enhance } from '$app/forms';
	import ShieldKey from 'svelte-material-icons/ShieldKey.svelte';

	export let data: PageData;

	let form: HTMLFormElement;

	const webauthnLogin = async () => {
		try {
			const attestation = JSONtoPublicKeyCredentialRequestOptions(data.loginStartData.response);
			const credential = await navigator.credentials.get({ publicKey: attestation });

			const formData = new FormData(form);

			formData.set('username', data.username);
			formData.set('data', JSON.stringify(LoginPublicKeyCredentialToJSON(credential)));

			const response = await fetch(form.action, {
				method: form.method,
				body: formData
			}).then((res) => res.json());

			if (response?.type === 'redirect') {
				goto('/login/successful');
			} else {
				console.error(response);
				goto('/login/webauthn/failed?username' + data.username);
			}
		} catch (e) {
			console.error(e);
			goto('/login/webauthn/failed?username' + data.username);
		}
	};

	onMount(webauthnLogin);
</script>

<form bind:this={form} use:enhance method="POST">
	<input type="hidden" name="username" value={data.username} />
	<input type="hidden" name="data" />
</form>

<h1 class="text-2xl font-bold">Login using a Passkey</h1>
<div class="label flex flex-col justify-center items-center">
	<div class="w-max h-max relative">
		<div class="loading loading-ring w-32" />
		<div class="absolute top-[50%] left-[50%] -translate-x-1/2 -translate-y-1/2">
			<!-- <i class="material-symbols-outlined text-5xl">passkey</i> -->
			<ShieldKey class="text-5xl" />
		</div>
	</div>
</div>

<h1 class="text-sm font-bold">
	Please insert your passkey into your device and press the button on it
</h1>

<div class="label mt-8">
	<a href="/login" class="label-text-alt link link-hover">
		Click here to go back to the login page
	</a>
</div>
