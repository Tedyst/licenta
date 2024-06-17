<script lang="ts">
	import { webauthnRegisterBegin, webauthnRegisterFinish } from '$lib/client';
	import {
		JSONtoPublicKeyCredentialCreationOptions,
		PublicKeyCredentialToJSON
	} from '$lib/webauthn';

	let error: string | null = null;

	const registerWebauthn = async () => {
		const registerData = await webauthnRegisterBegin().catch((e) => {
			error = e.message;
		});
		if (!registerData) {
			return;
		}
		const options = JSONtoPublicKeyCredentialCreationOptions(registerData.response);
		const credential = await navigator.credentials.create({ publicKey: options }).catch((e) => {
			error = e.message;
		});
		if (!credential) {
			return;
		}
		const credentialJSON = PublicKeyCredentialToJSON(credential);
		const registerResponse = await webauthnRegisterFinish(JSON.stringify(credentialJSON)).catch(
			(e) => {
				error = e.message;
			}
		);
	};

	// const testLogin = async () => {
	// 	const data = await fetchWebauthnLogin();

	// 	console.log(data);

	// 	const asd = {
	// 		publicKey: data.response
	// 	};

	// 	asd.publicKey.challenge = Base64Binary.decode(asd.publicKey.challenge, null);
	// 	asd.publicKey?.allowCredentials?.forEach((cred: any) => {
	// 		cred.id = Base64Binary.decode(cred.id, null);
	// 	});

	// 	console.log(asd);
	// 	let credential: any = await navigator.credentials.get(asd);

	// 	console.log(credential);
	// 	if (!credential) {
	// 		return;
	// 	}
	// 	fetch('/api/auth/webauthn/login/finish', {
	// 		method: 'POST',
	// 		headers: {
	// 			'Content-Type': 'application/json',
	// 			'X-CSRF-Token': await getCSRFToken()
	// 		},
	// 		body: JSON.stringify({
	// 			id: credential.id,
	// 			rawId: Base64Binary.encode(credential.rawId),
	// 			type: credential.type,
	// 			response: {
	// 				authenticatorData: Base64Binary.encode(credential.response.authenticatorData),
	// 				clientDataJSON: Base64Binary.encode(credential.response.clientDataJSON),
	// 				signature: Base64Binary.encode(credential.response.signature),
	// 				userHandle: Base64Binary.encode(credential.response.userHandle)
	// 			}
	// 		})
	// 	});
	// };
</script>

<button class="btn btn-primary" on:click={registerWebauthn}>Register</button>
<br />
{#if error}
	<div role="alert" class="alert alert-error">
		<svg
			xmlns="http://www.w3.org/2000/svg"
			class="stroke-current shrink-0 h-6 w-6"
			fill="none"
			viewBox="0 0 24 24"
			><path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
			/></svg
		>
		<span>{error}</span>
	</div>
{/if}
<!-- <button on:click={testLogin}>Login</button> -->
