<script lang="ts">
	import { goto } from '$app/navigation';
	import LoginSuccessful from '$lib/login/login-successful.svelte';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	onMount(async () => {
		await new Promise((resolve) => setTimeout(resolve, 2000));
		if ($page.url.pathname === '/login/successful') goto('/');
	});
</script>

<LoginSuccessful />
