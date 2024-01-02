<script lang="ts">
	import { username, validateUsername } from '$lib/login/login';
	import Login from '$lib/login/login.svelte';
	import { goto } from '$app/navigation';

	let error: string | null = null;
	const validate = (e: SubmitEvent) => {
		error = validateUsername($username);

		if (!error) {
			goto('/login/password');
		}
	};
</script>

<Login
	bind:username={$username}
	on:submit={validate}
	on:input={() => {
		error = null;
	}}
	{error}
/>
