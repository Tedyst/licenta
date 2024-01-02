<script lang="ts">
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

	const onSubmit = (e: SubmitEvent) => {
		e.preventDefault();
		const form = e.target as HTMLFormElement;
		const formData = new FormData(form);
		errors.username = validateUsername(formData.get('username') as string);
		errors.email = validateEmail(formData.get('email') as string);
		errors.password = validatePassword(formData.get('password') as string);

		if (errors.username || errors.email || errors.password) {
			return;
		}

		console.log(e);
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
