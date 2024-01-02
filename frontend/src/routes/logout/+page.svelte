<script lang="ts">
	import { goto } from '$app/navigation';
	import { logout } from '$lib/client';
	import { onMount } from 'svelte';

	onMount(() => {
		logout().then(() => {
			goto('/login');
		});
	});
</script>

<p>Logging out...</p>
