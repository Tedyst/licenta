<script lang="ts">
	import { enhance } from '$app/forms';
	import type { ActionData } from './$types';

	export let form: ActionData;
</script>

<form use:enhance method="POST">
	<div class="form-control">
		<label class="label" for="token">
			<span class="label-text">2FA Token</span>
		</label>
		<input
			type="text"
			placeholder="token"
			class="input input-bordered {form?.error
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="token"
			name="token"
		/>
		{#if form?.error}
			<div class="label text-error text-xs">
				{form?.error}
			</div>
		{/if}

		<div class="label">
			<a class="label-text-alt link link-hover" href="/login">
				Click here to go back to the login page
			</a>
		</div>
	</div>
	<div class="form-control mt-6">
		<button
			class="btn {!form?.error
				? 'btn-primary'
				: 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit">Login</button
		>
	</div>
</form>
