<script lang="ts">
	export let username: string | null;
	export let error: string | null;
	export let loading: boolean = false;
	export let sent: boolean = false;

	let color = 'btn-primary' as 'btn-primary' | 'btn-error' | 'btn-success';
	if (sent) {
		color = 'btn-success';
	}
	if (error) {
		color = 'btn-error';
	}
</script>

<form on:submit|preventDefault on:input>
	<div class="form-control">
		<label class="label" for="username">
			<span class="label-text">Username</span>
		</label>
		<input
			type="text"
			placeholder="username"
			class="input input-bordered {error
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="username"
			name="username"
			bind:value={username}
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
		<button class="btn {color} transition-colors duration-300 ease-in-out" type="submit">
			{#if loading && !error}
				<span class="loading loading-spinner" />
			{/if}
			{#if sent && !error}
				Password reset email sent
			{:else}
				Request password reset
			{/if}
		</button>
	</div>
</form>
