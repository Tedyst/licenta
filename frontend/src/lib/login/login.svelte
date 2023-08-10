<script lang="ts">
	export let errors: {
		password: string | null;
		email: string | null;
	} = {
		password: null,
		email: null
	};
	export let loading: boolean = false;
</script>

<form on:submit|preventDefault on:input={() => (errors = { password: null, email: null })}>
	<div class="form-control">
		<label class="label" for="email">
			<span class="label-text">Username or Email</span>
		</label>
		<input
			type="text"
			placeholder="username/email"
			class="input input-bordered {errors.email
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="email"
			name="email"
			autocomplete="email"
		/>
	</div>
	<div class="form-control">
		<label class="label" for="password">
			<span class="label-text">Password</span>
		</label>
		<input
			type="password"
			placeholder="password"
			class="input input-bordered {errors.password
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="password"
			name="password"
			autocomplete="current-password"
		/>
		{#if errors.password}
			<div class="label text-error text-xs">
				{errors.password}
			</div>
		{/if}

		<div class="label">
			<a href="/login/forgot-password" class="label-text-alt link link-hover">Forgot password?</a>
		</div>
	</div>
	<div class="form-control mt-6">
		<button
			class="btn {!errors.password && !errors.email
				? 'btn-primary'
				: 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit"
		>
			{#if loading}
				<span class="loading loading-spinner" />
			{/if}
			Login
		</button>
	</div>
</form>
