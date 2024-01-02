<script lang="ts">
	import { validateUsername } from '$lib/login/login';
	import { goto } from '$app/navigation';
	import ForgotPassword from '$lib/login/forgot-password.svelte';
	import { username as usernameStore } from '$lib/login/login';

	let loading = false;
	let sent = false;
	let errors: {
		username: string | null;
	} = {
		username: null
	};

	let onSubmit = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);
		let username = formData.get('username');
		if (typeof username !== 'string') {
			throw new Error('username must be a string');
		}

		errors = {
			username: validateUsername(username)
		};
		if (errors.username) {
			return;
		}

		loading = true;

		setTimeout(() => {
			sent = true;
			loading = false;
		}, 1000);

		setTimeout(() => {
			goto('/login');
		}, 3000);
	};
</script>

<ForgotPassword {errors} on:submit={onSubmit} {loading} {sent} bind:username={$usernameStore} />
