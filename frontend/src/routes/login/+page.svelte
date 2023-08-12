<script lang="ts">
	import { quartInOut } from 'svelte/easing';
	import { flyabsolute } from '$lib/animations';
	import { validatePassword, validateUsername } from '$lib/login/login';
	import { username } from '$lib/login/login';
	import { user } from '$lib/stores';

	import Login from '$lib/login/login.svelte';
	import { goto } from '$app/navigation';

	let loading = false;

	let errors: {
		username: string | null;
		password: string | null;
	};

	let onSubmit = (data: { username: string; password: string }) => {
		errors = {
			password: validatePassword(data.password),
			username: validateUsername(data.username)
		};
		if (errors.password || errors.username) {
			return;
		}

		loading = true;

		fetch('/api/v1/login', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(data)
		})
			.catch((err) => {
				console.error(err);
				errors = {
					username: null,
					password: 'An error occurred'
				};
			})
			.then((result) => result?.json())
			.then((data) => {
				if (!data?.success) {
					if (data?.message === '2fa required') {
						return;
					}
					goto('/goto/2fa');
					errors = {
						username: null,
						password: data?.message || 'An error occurred'
					};
					return;
				}
				$user = data?.user;
				goto('/dashboard');
			})
			.finally(() => {
				loading = false;
			});
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
	<Login {onSubmit} bind:errors bind:loading bind:username={$username} />
</div>
