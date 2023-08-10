<script lang="ts">
	import { cubicOut, quartInOut } from 'svelte/easing';
	import { fly, type TransitionConfig } from 'svelte/transition';

	import Login from '$lib/login.svelte';
	import Login2fa from '$lib/login2fa.svelte';
	import { onMount } from 'svelte';

	let need2fa = false;

	let loginErrors: {
		email?: string;
		password?: string;
	} = {};

	let twofaErrors: {
		token?: string;
	} = {};

	function goBackToLoginPage() {
		need2fa = false;
		twofaErrors = {};
		loginErrors = {};
	}

	function validatePassword(password: string) {
		return '';
		if (password.length < 8) {
			return 'Password must be at least 8 characters long';
		}
		return '';
	}

	let onLoginSubmit = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);

		let email = formData.get('email');
		let password = formData.get('password');

		if (typeof email !== 'string') {
			throw new Error('Email must be a string');
		}
		if (typeof password !== 'string') {
			throw new Error('Password must be a string');
		}

		if (validatePassword(password)) {
			loginErrors['password'] = validatePassword(password);
			return;
		}

		need2fa = true;
	};

	let on2faSubmit = (e: SubmitEvent) => {
		const formData = new FormData(e.target as HTMLFormElement);

		let token = formData.get('token');

		if (typeof token !== 'string') {
			throw new Error('Token must be a string');
		}

		if (token.length !== 6) {
			twofaErrors['token'] = 'Token must be 6 characters long';
			return;
		}
	};

	function makeabsolute(
		node: Element,
		{ delay = 0, duration = 400, easing = cubicOut, x = 0, y = 0, opacity = 0 } = {}
	): TransitionConfig {
		const flyConfig = fly(node, { delay, duration, easing, x, y, opacity });
		return {
			...flyConfig,
			css: (t, u) =>
				`opacity: ${
					t * u
				}; position: absolute; margin-left: 0; margin-right: 0; left: 0; right: 0; text-align: center; padding: 2rem; ${flyConfig?.css?.(
					t,
					u
				)};`
		};
	}
</script>

<div class="card flex-shrink-0 w-full max-w-sm shadow-2xl bg-base-100">
	<div class="card-body">
		{#if need2fa}
			<div
				in:makeabsolute={{ delay: 0, duration: 500, easing: quartInOut, x: 300 }}
				out:makeabsolute={{ delay: 0, duration: 500, easing: quartInOut, x: 300 }}
			>
				<Login2fa errors={twofaErrors} on:submit={on2faSubmit} {goBackToLoginPage} />
			</div>
		{:else}
			<div
				in:fly={{ delay: 0, duration: 500, easing: quartInOut, x: -300 }}
				out:fly={{ duration: 500, easing: quartInOut, x: -300 }}
			>
				<Login on:submit={onLoginSubmit} bind:errors={loginErrors} />
			</div>
		{/if}
	</div>
</div>
