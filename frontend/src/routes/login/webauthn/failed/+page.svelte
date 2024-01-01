<script>
	import { flyabsolute } from '$lib/animations';
	import LoginWebauthnFailed from '$lib/login/login-webauthn-failed.svelte';
	import { quartInOut } from 'svelte/easing';
</script>

<!-- <div
	in:flyabsolute={{
		delay: 0,
		duration: 500,
		easing: quartInOut,
		x: 300,
		otherStyling: 'text-align: center; padding: 2rem;'
	}}
	out:flyabsolute={{
		delay: 0,
		duration: 500,
		easing: quartInOut,
		x: 300,
		otherStyling: 'text-align: center; padding: 2rem;'
	}}
> -->
<LoginWebauthnFailed />
<!-- </div> -->
