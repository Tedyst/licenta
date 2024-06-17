<script lang="ts">
	import type { LayoutData } from './$types';
	export let data: LayoutData;

	import MainLayout from '$lib/utils/MainLayout.svelte';
</script>

<MainLayout organizations={data.organizations} user={data.user} serverError={data.error}>
	<slot />
</MainLayout>
