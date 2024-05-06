<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
</script>

<div>
	<h1>Dashboard</h1>

	<p>Logged in as {data.user.username}</p>
</div>

<svelte:head>
	<title>Dashboard | Licenta</title>
</svelte:head>
