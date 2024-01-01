<script lang="ts">
	export let username: string;
	export let error: string | null = null;
	export let loading: boolean = false;
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
		/>
	</div>
	<div class="form-control">
		<div class="label">
			<a href="/login/webauthn" class="label-text-alt link link-hover"> Sign in using a passkey </a>
		</div>
	</div>
	<div class="form-control mt-6">
		<button
			class="btn {!error ? 'btn-primary' : 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit"
		>
			{#if loading}
				<span class="loading loading-spinner" />
			{/if}
			Login
		</button>
	</div>
</form>
