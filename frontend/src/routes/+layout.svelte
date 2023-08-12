<script lang="ts">
	import '../app.css';
	import { theme } from '../lib/stores';
	import { browser } from '$app/environment';

	$: if (browser) {
		document.documentElement.setAttribute('data-theme', $theme);
	}
</script>

<div class="min-h-screen bg-base-200">
	<slot />
</div>

<svelte:head>
	<script>
		(function () {
			let localTheme = localStorage.getItem('theme');
			if (localTheme) {
				if (typeof document === 'undefined') return;
				document.documentElement.setAttribute('data-theme', JSON.parse(localTheme));
			}
		})();
	</script>
</svelte:head>
