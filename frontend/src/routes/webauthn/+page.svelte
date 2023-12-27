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

	const test = async () => {
		const data = await fetchWebauthnRegistration();

		console.log(data);

		const asd = {
			publicKey: data.response
		};

		asd.publicKey.challenge = Base64Binary.decode(asd.publicKey.challenge, null);
		asd.publicKey.user.id = Base64Binary.decode(asd.publicKey.user.id, null);

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
</script>

<button on:click={test}>Register</button>
