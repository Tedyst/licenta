<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';

	export let form: ActionData;
	export let data: PageData;

	let username = data?.username ?? '';
</script>

<form method="POST" use:enhance action="?/password">
	<div class="form-control">
		{#if form?.error}
			<label class="label" for="username">
				<span class="label-text text-error">{form.error}</span>
			</label>
		{:else}
			<label class="label" for="username">
				<span class="label-text">Username</span>
			</label>
		{/if}
		<!-- svelte-ignore a11y-autofocus -->
		<input
			type="text"
			placeholder="Username"
			class="input input-bordered {form?.error
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="username"
			name="username"
			autocomplete="username"
			bind:value={username}
			autofocus
		/>
	</div>
	<div class="form-control mt-4">
		<div class="label">
			<form method="POST" use:enhance action="?/webauthn">
				<input type="hidden" name="username" value={username} />
				<button type="submit" class="label-text-alt link link-hover">Sign in using a passkey</button
				>
			</form>
		</div>
		<div class="label pt-0 mt-0">
			<a href="/register" class="label-text-alt link link-hover"> Register an account </a>
		</div>
	</div>
	<div class="form-control mt-6">
		<button
			class="btn {!form?.error
				? 'btn-primary'
				: 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit"
		>
			Login
		</button>
	</div>
</form>
