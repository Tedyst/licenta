<script lang="ts">
	import type { ActionData } from './$types';
	import { enhance } from '$app/forms';

	export let form: ActionData;
</script>

<form method="POST" use:enhance>
	<div class="form-control mt-2">
		{#if form?.usernameError}
			<label class="label" for="username">
				<span class="label-text text-error">{form?.usernameError}</span>
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
			class="input input-bordered {form?.usernameError
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="username"
			name="username"
			autocomplete="username"
			autofocus
		/>
	</div>
	<div class="form-control mt-2">
		{#if form?.emailError}
			<label class="label" for="email">
				<span class="label-text text-error">{form?.emailError}</span>
			</label>
		{:else}
			<label class="label" for="email">
				<span class="label-text">Email</span>
			</label>
		{/if}
		<input
			type="text"
			placeholder="Email"
			class="input input-bordered {form?.emailError
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="email"
			name="email"
			autocomplete="email"
		/>
	</div>
	<div class="form-control mt-2">
		{#if form?.passwordError}
			<label class="label" for="password">
				<span class="label-text text-error">{form?.passwordError}</span>
			</label>
		{:else}
			<label class="label" for="password">
				<span class="label-text">Password</span>
			</label>
		{/if}
		<input
			type="password"
			placeholder="Password"
			class="input input-bordered {form?.passwordError
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="password"
			name="password"
		/>
	</div>
	<div class="form-control mt-6">
		<button
			class="btn {!form?.usernameError && !form?.emailError && !form?.passwordError && !form?.error
				? 'btn-primary'
				: 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit"
		>
			{#if form?.error}
				{form.error}
			{:else}
				Register
			{/if}
		</button>
	</div>
</form>
