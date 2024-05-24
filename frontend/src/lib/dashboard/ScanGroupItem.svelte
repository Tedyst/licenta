<script lang="ts">
	import type { components } from '$lib/api/v1';

	export let scanGroup: components['schemas']['ScanGroup'];
</script>

<div
	class="card bg-base-100 text-lg font-bold flex flex-col gap-3 shadow-xl grow place-content-around mt-1 mb-1"
>
	<div class="card-body flex-col lg:flex-row">
		<div class="flex flex-col gap-3 grow">
			<h2 class="overflow-auto break-all">Scan Group {scanGroup.id}</h2>
			<div class="flex flex- gap-3">
				<slot />
			</div>
		</div>
	</div>
</div>
