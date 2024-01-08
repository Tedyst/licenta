<script lang="ts">
	import '../app.css';
	import { theme } from '../lib/stores';
	import { browser } from '$app/environment';
	import { PlausibleAnalytics } from '@accuser/svelte-plausible-analytics';

	$: if (browser) {
		document.documentElement.setAttribute('data-theme', $theme);
	}
</script>

<PlausibleAnalytics
	apiHost="https://plausible.tedyst.ro"
	domain="laptop.tedyst.ro"
	enabled={true}
/>

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
