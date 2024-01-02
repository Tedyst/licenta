<script lang="ts">
	import { quartInOut } from 'svelte/easing';
	import { flyabsolute } from '$lib/animations';
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
				error = res?.errors?.code?.at(0) || res?.message || 'Unknown error';
			})
			.catch((err) => {
				error = err.message;
			});
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
	<Login2fa bind:error on:submit={onSubmit} />
</div>
