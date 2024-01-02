<script lang="ts">
	export let error: string;
	export let loading: boolean = false;
</script>

<form on:submit|preventDefault on:input={() => (error = '')}>
	<div class="form-control">
		<label class="label" for="token">
			<span class="label-text">2FA Token</span>
		</label>
		<input
			type="number"
			placeholder="token"
			class="input input-bordered {error
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="token"
			name="token"
		/>
		{#if error}
			<div class="label text-error text-xs">
				{error}
			</div>
		{/if}

		<div class="label">
			<a class="label-text-alt link link-hover" href="/login">
				Click here to go back to the login page
			</a>
		</div>
	</div>
	<div class="form-control mt-6">
		{#if loading}
			<span class="loading loading-spinner" />
		{/if}
		<button
			class="btn {!error ? 'btn-primary' : 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit">Login</button
		>
	</div>
</form>
