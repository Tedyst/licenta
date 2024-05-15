<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { page } from '$app/stores';

	onMount(async () => {
		await new Promise((resolve) => setTimeout(resolve, 2000));
		if ($page.url.pathname === '/login/successful') goto('/');
	});
</script>

<h1 class="text-2xl font-bold">You are now logged in</h1>

<h1 class="text-sm">Redirecting in 2 seconds...</h1>

<h1 class="text-sm">If you are not redirected, click <a href="/">here</a></h1>
