<script lang="ts">
	import { quartInOut } from 'svelte/easing';
	import { flyabsolute } from '$lib/animations';
	import { username, validateUsername } from '$lib/login/login';

	import Login from '$lib/login/login.svelte';

	let error: string | null = null;
	const validate = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);
		let username = formData.get('username');

		if (!username) {
			return 'Please enter a username';
		}
		if (validateUsername(username as string)) {
			error = validateUsername(username as string);
		}

		console.log(error);

		if (error) return e.preventDefault();
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
