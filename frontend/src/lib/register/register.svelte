<script lang="ts">
	import { onMount } from 'svelte';

	export let errors: {
		username: string | null;
		email: string | null;
		password: string | null;
	};

	let elm: HTMLInputElement;

	onMount(() => {
		elm.focus();
	});
</script>

<form on:submit|preventDefault on:input>
	<div class="form-control mt-2">
		{#if errors?.username}
			<label class="label" for="username">
				<span class="label-text text-error">{errors?.username}</span>
			</label>
		{:else}
			<label class="label" for="username">
				<span class="label-text">Username</span>
			</label>
		{/if}
		<input
			type="text"
			placeholder="Username"
			class="input input-bordered {errors.username
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="username"
			name="username"
			autocomplete="username"
			bind:this={elm}
		/>
	</div>
	<div class="form-control mt-2">
		{#if errors?.email}
			<label class="label" for="email">
				<span class="label-text text-error">{errors?.email}</span>
			</label>
		{:else}
			<label class="label" for="email">
				<span class="label-text">Email</span>
			</label>
		{/if}
		<input
			type="text"
			placeholder="Email"
			class="input input-bordered {errors.email
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="email"
			name="email"
			autocomplete="email"
			bind:this={elm}
		/>
	</div>
	<div class="form-control mt-2">
		{#if errors?.password}
			<label class="label" for="password">
				<span class="label-text text-error">{errors?.password}</span>
			</label>
		{:else}
			<label class="label" for="password">
				<span class="label-text">Password</span>
			</label>
		{/if}
		<input
			type="password"
			placeholder="Password"
			class="input input-bordered {errors.password
				? 'wiggle input-error'
				: ''} transition-colors duration-300 ease-in-out"
			id="password"
			name="password"
			bind:this={elm}
		/>
	</div>
	<div class="form-control mt-6">
		<button
			class="btn {!errors?.password && !errors?.username && !errors?.email
				? 'btn-primary'
				: 'btn-error'} transition-colors duration-300 ease-in-out"
			type="submit"
		>
			Register
		</button>
	</div>
</form>
