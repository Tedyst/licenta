<script lang="ts">
	import LoginPassword from '$lib/login/login-password.svelte';
	import { username } from '$lib/login/login';
	import { goto } from '$app/navigation';
	import { login } from '$lib/client';
	import { onMount } from 'svelte';

	let error: string | null = null;

	onMount(() => {
		if ($username === null) {
			goto('/login');
		}
	});

	const submit = async (e: SubmitEvent) => {
		e.preventDefault();
		if ($username === null) {
			return;
		}
		const form = e.target as HTMLFormElement;
		const formData = new FormData(form);
		const password = formData.get('password') as string;
		const remember = formData.get('remember') as string;
		const loginResponse = await login($username, password, remember == 'on');
		console.log(loginResponse);
		if (loginResponse.success) {
			goto('/login/successful');
		} else if (loginResponse.totp && loginResponse.webauthn) {
			goto('/login/2fa');
		} else if (loginResponse.totp) {
			goto('/login/totp');
		} else if (loginResponse.webauthn) {
			goto('/login/webauthn');
		} else {
			error = loginResponse.error || null;
		}
	};
</script>

<LoginPassword username={$username} on:submit={submit} bind:error />
