<script lang="ts">
	export let username: string;
	export let errors: {
		password: string | null;
		username: string | null;
	} = {
		password: null,
		username: null
	};
	export let loading: boolean = false;
	export let onSubmit: (data: { username: string; password: string }) => void;

	const submit = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);
		let username = formData.get('username');
		let password = formData.get('password');
		if (typeof username !== 'string') {
			throw new Error('username must be a string');
		}
		if (typeof password !== 'string') {
			throw new Error('Password must be a string');
		}

		onSubmit({ username, password });
	};
</script>

<form
	on:submit|preventDefault={submit}
	on:input={() => (errors = { password: null, username: null })}
>
	<div class="form-control">
		<label class="label" for="username">
			<span class="label-text">Username or username</span>
		</label>
		<input
			type="text"
			placeholder="username/username"
			class="input input-bordered {errors.username
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="username"
			name="username"
			autocomplete="username"
			bind:value={username}
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

		<div class="label">
			<a href="/login/webauthn" class="label-text-alt link link-hover">
				Sign in using a security key
			</a>
		</div>
	</div>
	<div class="form-control mt-6">
		<button
			class="btn {!errors.password && !errors.username
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
