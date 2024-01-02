<script lang="ts">
	import { goto } from '$app/navigation';
	import ForgotPassword from '$lib/login/forgot-password.svelte';
	import { username } from '$lib/login/login';
	import { requestResetPassword } from '$lib/client';

	let error: string | null = null;

	let promise: Promise<void> | null = null;

	let onSubmit = (e: SubmitEvent) => {
		if ($username === null) {
			return;
		}
		promise = requestResetPassword($username)
			.then((response) => {
				if (response.success) {
					setTimeout(() => {
						goto('/login');
					}, 2000);
				}
			})
			.catch((e) => {
				console.log(e);
				error = e.message;
			});
	};
</script>

{#if promise == null}
	<ForgotPassword
		{error}
		on:submit={onSubmit}
		loading={false}
		sent={false}
		bind:username={$username}
	/>
{:else}
	{#await promise}
		<ForgotPassword
			{error}
			on:submit={onSubmit}
			loading={true}
			sent={false}
			bind:username={$username}
		/>
	{:then}
		<ForgotPassword
			{error}
			on:submit={onSubmit}
			loading={false}
			sent={true}
			bind:username={$username}
		/>
	{/await}
{/if}
