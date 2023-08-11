<script lang="ts">
	import { quartInOut } from 'svelte/easing';
	import { flyabsolute } from '$lib/animations';
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

<div
	in:flyabsolute={{
		delay: 0,
		duration: 500,
		easing: quartInOut,
		x: 300,
		otherStyling: 'text-align: center; padding: 2rem;'
	}}
	out:flyabsolute={{
		delay: 0,
		duration: 500,
		easing: quartInOut,
		x: 300,
		otherStyling: 'text-align: center; padding: 2rem;'
	}}
>
	<ForgotPassword {errors} on:submit={onSubmit} {loading} {sent} bind:username={$usernameStore} />
</div>
