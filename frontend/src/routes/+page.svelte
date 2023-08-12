<script lang="ts">
	import { theme } from '../lib/stores';
	import { user } from '$lib/stores';

	function toggleTheme() {
		theme.set($theme === 'dark' ? 'light' : 'dark');
	}
</script>

<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://kit.svelte.dev">kit.svelte.dev</a> to read the documentation</p>

<button class="btn btn-secondary" on:click={toggleTheme}>Secondary</button>

{#if $user}
	<a href="/dashboard" class="btn btn-primary">Go to Dashboard</a>
{:else}
	<a href="/login" class="btn btn-primary">Go to Dashboard</a>
{/if}

<svelte:head>
	<title>SvelteKit</title>
</svelte:head>
