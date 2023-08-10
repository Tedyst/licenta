<script lang="ts">
	const email: string = '';
	const password: string = '';
	export let errors: {
		password?: string;
		email?: string;
	} = {};

	function clearErrors() {
		errors = {};
	}
</script>

<form on:submit|preventDefault on:input={clearErrors}>
	<div class="form-control">
		<label class="label" for="email">
			<span class="label-text">Email</span>
		</label>
		<input
			type="email"
			placeholder="email"
			class="input input-bordered {errors.email
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="email"
			name="email"
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
		/>
		{#if errors.password}
			<div class="label text-error text-xs">
				{errors.password}
			</div>
		{/if}

		<div class="label">
			<a href="#/" class="label-text-alt link link-hover">Forgot password?</a>
		</div>
	</div>
	<div class="form-control mt-6">
		<button
			class="btn {Object.keys(errors).length === 0
				? 'btn-primary'
				: 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit">Login</button
		>
	</div>
</form>
