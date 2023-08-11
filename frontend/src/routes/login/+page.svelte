<script lang="ts">
	import { quartInOut } from 'svelte/easing';
	import { flyabsolute } from '$lib/animations';
	import { validatePassword, validateUsername } from '$lib/login/login';
	import { username } from '$lib/login/login';

	import Login from '$lib/login/login.svelte';
	import { goto } from '$app/navigation';

	let loading = false;

	let errors: {
		email: string | null;
		password: string | null;
	};

	let onSubmit = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);
		let email = formData.get('email');
		let password = formData.get('password');
		if (typeof email !== 'string') {
			throw new Error('Email must be a string');
		}
		if (typeof password !== 'string') {
			throw new Error('Password must be a string');
		}

		errors = {
			password: validatePassword(password),
			email: validateUsername(email)
		};
		if (errors.password || errors.email) {
			return;
		}

		loading = true;

		setTimeout(() => {
			goto('/login/totp');
		}, 1000);
	};
</script>

<div
	in:flyabsolute={{
		delay: 0,
		duration: 500,
		easing: quartInOut,
		x: -300,
		otherStyling: 'text-align: center; padding: 2rem;'
	}}
	out:flyabsolute={{
		duration: 500,
		easing: quartInOut,
		x: -300,
		otherStyling: 'text-align: center; padding: 2rem;'
	}}
>
	<Login on:submit={onSubmit} bind:errors bind:loading bind:username={$username} />
</div>
