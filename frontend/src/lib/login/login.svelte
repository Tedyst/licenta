<script lang="ts">
	export let username: string | null = null;
	export let error: string | null = null;
	import { onMount } from 'svelte';

	let elm: HTMLInputElement;

	onMount(() => {
		elm.focus();
	});
</script>

<form on:submit|preventDefault on:input>
	<div class="form-control">
		{#if error}
			<label class="label" for="username">
				<span class="label-text text-error">{error}</span>
			</label>
		{:else}
			<label class="label" for="username">
				<span class="label-text">Username</span>
			</label>
		{/if}
		<input
			type="text"
			placeholder="Username"
			class="input input-bordered {error
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="username"
			name="username"
			autocomplete="username"
			bind:value={username}
			bind:this={elm}
		/>
	</div>
	<div class="form-control mt-4">
		<div class="label">
			<a href="/login/webauthn" class="label-text-alt link link-hover"> Sign in using a passkey </a>
		</div>
		<div class="label pt-0 mt-0">
			<a href="/register" class="label-text-alt link link-hover"> Register an account </a>
		</div>
	</div>
	<div class="form-control mt-6">
		<button
			class="btn {!error ? 'btn-primary' : 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit"
		>
			Login
		</button>
	</div>
</form>
