<script lang="ts">
	import { goto } from '$app/navigation';
	import Skeleton from '$lib/dashboard/skeleton.svelte';
	import { user } from '$lib/stores';
	import { onMount } from 'svelte';

	var promise: Promise<any> | null = null;
	onMount(() => {
		if (!$user) {
			promise = fetch('/api/v1/users/me')
				.then((result) => result?.json())
				.then((data) => {
					if (!data?.success) {
						goto('/login');
						return;
					}
					$user = data.user;
				});
		}
	});
</script>

<svelte:head>
	<script>
		(function () {
			let localTheme = localStorage.getItem('user');
			if (localTheme) {
				if (typeof document === 'undefined') return;
				$user = JSON.parse(localTheme);
			}
		})();
	</script>
</svelte:head>

{#if !$user}
	<Skeleton />
{:else}
	<slot />
{/if}
