<script lang="ts">
	import { quartInOut } from 'svelte/easing';
	import { flyabsolute } from '$lib/animations';
	import { username, validateUsername } from '$lib/login/login';

	import Login from '$lib/login/login.svelte';
	import { goto } from '$app/navigation';

	let error: string | null = null;
	const validate = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);
		let u = formData.get('username');

		if (!u) {
			return 'Please enter a username';
		}
		if (validateUsername(u as string)) {
			error = validateUsername(u as string);
		}

		console.log(error);

		if (!error) {
			$username = u as string;
			goto('/login/password');
		}
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
	<Login bind:username={$username} on:submit={validate} {error} />
</div>
