<script>
	import { flyabsolute } from '$lib/animations';
	import LoginWebauthn from '$lib/login/login-webauthn.svelte';
	import { quartInOut } from 'svelte/easing';
	import { username } from '$lib/login/login';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { webauthnLoginBegin, webauthnLoginFinish } from '$lib/client';
	import {
		JSONtoPublicKeyCredentialRequestOptions,
		LoginPublicKeyCredentialToJSON
	} from '$lib/webauthn';

	const webauthnLogin = async () => {
		const loginStartData = await webauthnLoginBegin($username);
		const attestation = JSONtoPublicKeyCredentialRequestOptions(loginStartData.response);
		const credential = await navigator.credentials.get({ publicKey: attestation });
		const sendData = LoginPublicKeyCredentialToJSON(credential);
		return await webauthnLoginFinish(JSON.stringify(sendData));
	};

	onMount(async () => {
		try {
			const response = await webauthnLogin();
			if (response.success) {
				goto('/login/successful');
			} else {
				goto('/login/webauthn/failed');
			}
		} catch (e) {
			goto('/login/webauthn/failed');
			console.log(e);
		}
	});
</script>

<LoginWebauthn />
