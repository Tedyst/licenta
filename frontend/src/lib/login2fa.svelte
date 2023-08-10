<script lang="ts">
	export let errors: {
		token: string | null;
	} = { token: null };
</script>

<form on:submit|preventDefault on:input={() => (errors = { token: null })}>
	<div class="form-control">
		<label class="label" for="token">
			<span class="label-text">2FA Token</span>
		</label>
		<input
			type="number"
			placeholder="token"
			class="input input-bordered {errors.token
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="token"
			name="token"
		/>
		{#if errors.token}
			<div class="label text-error text-xs">
				{errors.token}
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
			class="btn {!errors.token
				? 'btn-primary'
				: 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit">Login</button
		>
	</div>
</form>
