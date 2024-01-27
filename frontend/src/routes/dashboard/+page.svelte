<script lang="ts">
	import { user } from '$lib/stores';
</script>

<div>
	<h1>Dashboard</h1>

	<p>Logged in as {$user?.username}</p>
</div>

<svelte:head>
	<title>Dashboard | Licenta</title>
</svelte:head>
