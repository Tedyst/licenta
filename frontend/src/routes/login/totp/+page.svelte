<script lang="ts">
	import Login2fa from '$lib/login/login-totp.svelte';
	import { goto } from '$app/navigation';
	import { loginTOTP } from '$lib/client';

	let error = '';

	let onSubmit = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);
		let token = formData.get('token');
		if (typeof token !== 'string') {
			throw new Error('Token must be a string');
		}
		loginTOTP(token)
			.then((res) => {
				if (res.success) {
					goto('/login/successful');
				}
				error = res?.errors?.code?.at(0) || res?.message || '';
			})
			.catch((err) => {
				error = err.message;
			});
	};
</script>

<Login2fa bind:error on:submit={onSubmit} />
