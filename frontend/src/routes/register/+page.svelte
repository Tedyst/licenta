<script lang="ts">
	import { goto } from '$app/navigation';
	import { registerUser } from '$lib/client';
	import { validateEmail, validatePassword, validateUsername } from '$lib/login/login';
	import Register from '$lib/register/register.svelte';

	let errors: {
		username: string | null;
		email: string | null;
		password: string | null;
	} = {
		username: null,
		email: null,
		password: null
	};

	const onSubmit = async (e: SubmitEvent) => {
		e.preventDefault();
		const form = e.target as HTMLFormElement;
		const formData = new FormData(form);
		errors.username = validateUsername(formData.get('username') as string);
		errors.email = validateEmail(formData.get('email') as string);
		errors.password = validatePassword(formData.get('password') as string);

		if (errors.username || errors.email || errors.password) {
			return;
		}

		let response = await registerUser({
			username: formData.get('username') as string,
			email: formData.get('email') as string,
			password: formData.get('password') as string
		});
		if (response.success) {
			await goto('/login');
		} else {
			errors.username = response.errors?.username?.join(', ') || null;
			errors.email = response.errors?.email?.join(', ') || null;
			errors.password = response.errors?.password?.join(', ') || null;
		}
	};
</script>

<Register
	{errors}
	on:submit={onSubmit}
	on:input={() => {
		errors = {
			username: null,
			email: null,
			password: null
		};
	}}
/>
