<script lang="ts">
	import { quartInOut } from 'svelte/easing';
	import { flyabsolute } from '$lib/animations';
	import { validateTOTPToken } from '$lib/login/login';
	import Login2fa from '$lib/login/login-totp.svelte';
	import { goto } from '$app/navigation';

	let loading = false;
	let errors: {
		token: string | null;
	} = {
		token: null
	};

	let onSubmit = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);
		let token = formData.get('token');
		if (typeof token !== 'string') {
			throw new Error('Token must be a string');
		}

		errors = {
			token: validateTOTPToken(token)
		};
		if (errors.token) {
			return;
		}

		loading = true;

		setTimeout(() => {
			goto('/');
		}, 1000);
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
	<Login2fa {errors} on:submit={onSubmit} />
</div>
