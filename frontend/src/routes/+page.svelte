<script lang="ts">
	import { theme } from '../lib/stores';
	import { user } from '$lib/stores';

	import { pa } from '@accuser/svelte-plausible-analytics';

	const { addEvent } = pa;

	function toggleTheme() {
		addEvent('theme-toggle');
		theme.set($theme === 'dark' ? 'light' : 'dark');
	}
</script>

<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://kit.svelte.dev">kit.svelte.dev</a> to read the documentation</p>

<button class="btn btn-secondary" on:click={toggleTheme}>Theme toggle</button>
<a href="/login" class="btn btn-secondary">Login</a>

<a href="/dashboard" class="btn btn-primary">Go to Dashboard</a>

<svelte:head>
	<title>SvelteKit</title>
</svelte:head>
