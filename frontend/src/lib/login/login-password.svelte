<script lang="ts">
	export const error: string | null = null;
	export let username: string;
</script>

<form on:submit|preventDefault on:input>
	<div class="form-control">
		Logging in as <span class="label-text-alt text-lg">{username}</span>
	</div>
	<div class="form-control mt-5">
		{#if error}
			<label class="label" for="password">
				<span class="label-text text-error">{error}</span>
			</label>
		{:else}
			<label class="label" for="password">
				<span class="label-text">Password</span>
			</label>
		{/if}
		<input
			type="password"
			placeholder="password"
			class="input input-bordered {error
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="password"
			name="password"
			autocomplete="current-password"
		/>
	</div>
	<div class="label mt-3">
		<a href="/login/forgot-password" class="label-text-alt link link-hover">Forgot password?</a>
	</div>
	<div class="label mt-0 pt-0">
		<a href="/login" class="label-text-alt link link-hover"
			>Not you? Click here to go to the login page</a
		>
	</div>
	<div class="label mt-0 pt-0">
		<a href="/login/webauthn" class="label-text-alt link link-hover"
			>Login using security key associated with account</a
		>
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
