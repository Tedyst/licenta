<script lang="ts">
	import type { ActionData, PageData } from './$types';
	import { enhance } from '$app/forms';

	export let form: ActionData;
	export let data: PageData;
</script>

<form method="POST" use:enhance>
	<div class="form-control">
		{#if form?.message}
			<div class="label text-success text-xs">
				{form?.message}
			</div>
		{/if}
		<label class="label" for="username">
			<span class="label-text">Username</span>
		</label>
		<input
			type="text"
			placeholder="username"
			class="input input-bordered {form?.error
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="username"
			name="username"
			value={data?.username}
		/>
		{#if form?.error}
			<div class="label text-error text-xs">
				{form?.error}
			</div>
		{/if}

		<div class="label">
			<a class="label-text-alt link link-hover" href="/login?username={data.username}">
				Click here to go back to the login page
			</a>
		</div>
	</div>
	<div class="form-control mt-6">
		<button class="btn duration-300 ease-in-out btn-primary" type="submit">
			Request password reset
		</button>
	</div>
</form>
