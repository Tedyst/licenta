<script lang="ts">
	import type { PageData } from './$types';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		JSONtoPublicKeyCredentialRequestOptions,
		LoginPublicKeyCredentialToJSON
	} from '$lib/webauthn';
	import { enhance } from '$app/forms';

	export let data: PageData;

	let responseData: string = '';
	let form: HTMLFormElement;

	const webauthnLogin = async () => {
		try {
			const attestation = JSONtoPublicKeyCredentialRequestOptions(data.loginStartData.response);
			const credential = await navigator.credentials.get({ publicKey: attestation });
			responseData = JSON.stringify(LoginPublicKeyCredentialToJSON(credential));
			form.requestSubmit();
		} catch (e) {
			console.error(e);
			goto('/login/webauthn/failed');
		}
	};

	onMount(webauthnLogin);
</script>

<form bind:this={form} use:enhance method="POST">
	<input type="hidden" name="username" value={data.username} />
	<input type="hidden" name="data" value={responseData} />
</form>

<h1 class="text-2xl font-bold">Login using a Passkey</h1>
<div class="label flex flex-col justify-center items-center">
	<div class="w-max h-max relative">
		<div class="loading loading-ring w-32" />
		<div class="absolute top-[50%] left-[50%] -translate-x-1/2 -translate-y-1/2">
			<i class="material-symbols-outlined text-5xl">passkey</i>
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
