<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;

	import MainLayout from '$lib/utils/MainLayout.svelte';
</script>

<MainLayout user={data.user} organizations={data.organizations} serverError={data.error}>
	<slot />
</MainLayout>
