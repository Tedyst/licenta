<script lang="ts">
	import { Base64Binary } from '$lib/Base64-binary';

	const getCSRFToken = async () => {
		const response = await fetch('/api/auth/webauthn/register/begin', {
			method: 'OPTIONS',
			headers: {
				'Content-Type': 'application/json'
			}
		});
		const data = await response.headers.get('X-CSRF-Token');
		return data || '';
	};

	const fetchWebauthnRegistration = async () => {
		const response = await fetch('/api/auth/webauthn/register/begin', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'X-CSRF-Token': await getCSRFToken()
			}
		});
		const data = await response.json();
		return data;
	};

	const fetchWebauthnLogin = async () => {
		const response = await fetch('/api/auth/webauthn/login/begin', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'X-CSRF-Token': await getCSRFToken()
			},
			body: '{}'
		});
		const data = await response.json();
		return data;
	};

	const testRegister = async () => {
		const data = await fetchWebauthnRegistration();

		console.log(data);

		const asd = {
			publicKey: data.response
		};

		asd.publicKey.challenge = Base64Binary.decode(asd.publicKey.challenge, null);
		asd.publicKey.user.id = Base64Binary.decode(asd.publicKey.user.id, null);
		asd.publicKey?.excludeCredentials?.forEach((cred: any) => {
			cred.id = Base64Binary.decode(cred.id, null);
		});

		console.log(asd);
		let credential: any = await navigator.credentials.create(asd);

		console.log(credential);
		if (!credential) {
			return;
		}
		fetch('/api/auth/webauthn/register/finish', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'X-CSRF-Token': await getCSRFToken()
			},
			body: JSON.stringify({
				id: credential.id,
				rawId: Base64Binary.encode(credential.rawId),
				type: credential.type,
				response: {
					attestationObject: Base64Binary.encode(credential.response.attestationObject),
					clientDataJSON: Base64Binary.encode(credential.response.clientDataJSON)
				}
			})
		});
	};

	const testLogin = async () => {
		const data = await fetchWebauthnLogin();

		console.log(data);

		const asd = {
			publicKey: data.response
		};

		asd.publicKey.challenge = Base64Binary.decode(asd.publicKey.challenge, null);
		asd.publicKey?.allowCredentials?.forEach((cred: any) => {
			cred.id = Base64Binary.decode(cred.id, null);
		});

		console.log(asd);
		let credential: any = await navigator.credentials.get(asd);

		console.log(credential);
		if (!credential) {
			return;
		}
		fetch('/api/auth/webauthn/login/finish', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'X-CSRF-Token': await getCSRFToken()
			},
			body: JSON.stringify({
				id: credential.id,
				rawId: Base64Binary.encode(credential.rawId),
				type: credential.type,
				response: {
					authenticatorData: Base64Binary.encode(credential.response.authenticatorData),
					clientDataJSON: Base64Binary.encode(credential.response.clientDataJSON),
					signature: Base64Binary.encode(credential.response.signature),
					userHandle: Base64Binary.encode(credential.response.userHandle)
				}
			})
		});
	};
</script>

<button on:click={testRegister}>Register</button>
<button on:click={testLogin}>Login</button>
