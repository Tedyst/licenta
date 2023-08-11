<script lang="ts">
	export let username: string;
	export let errors: {
		username: string | null;
	};
	export let loading: boolean = false;
	export let sent: boolean = false;

	let color = 'btn-primary';
	$: if (errors.username) {
		color = 'btn-error';
	}
	$: if (sent) {
		color = 'btn-success';
	}
</script>

<form on:submit|preventDefault on:input={() => (errors = { username: null })}>
	<div class="form-control">
		<label class="label" for="username">
			<span class="label-text">Username or Email</span>
		</label>
		<input
			type="text"
			placeholder="username/email"
			class="input input-bordered {errors.username
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="username"
			name="username"
			bind:value={username}
		/>
		{#if errors.username}
			<div class="label text-error text-xs">
				{errors.username}
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
			{#if loading}
				<span class="loading loading-spinner" />
			{/if}
			{#if sent}
				Password reset email sent
			{:else}
				Request password reset
			{/if}
		</button>
	</div>
</form>
