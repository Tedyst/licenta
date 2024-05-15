<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';

	export let form: ActionData;
	export let data: PageData;
</script>

<form method="POST" use:enhance>
	<div class="form-control">
		Logging in as <span class="label-text-alt text-lg font-bold">{data.username}</span>
	</div>
	<div class="form-control mt-5">
		{#if form?.error}
			<label class="label" for="password">
				<span class="label-text text-error">{form?.error}</span>
			</label>
		{:else}
			<label class="label" for="password">
				<span class="label-text">Password</span>
			</label>
		{/if}
		<!-- svelte-ignore a11y-autofocus -->
		<input
			type="password"
			placeholder="password"
			class="input input-bordered {form?.error
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="password"
			name="password"
			autocomplete="current-password"
			autofocus
		/>
		<input type="hidden" name="username" value={data.username} />
		<div class="form-control">
			<label class="label cursor-pointer">
				<span class="label-text">Remember me</span>
				<input type="checkbox" checked={true} class="checkbox" id="remember" name="remember" />
			</label>
		</div>
	</div>
	<div class="label mt-3">
		<a href="/login/forgot-password?username={data.username}" class="label-text-alt link link-hover"
			>Forgot password?</a
		>
	</div>
	<div class="label mt-0 pt-0">
		<a href="/login" class="label-text-alt link link-hover"
			>Not you? Click here to go to the login page</a
		>
	</div>
	<div class="label mt-0 pt-0">
		<a href="/login/webauthn" class="label-text-alt link link-hover"
			>Login using a passkey associated with account</a
		>
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
