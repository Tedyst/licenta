<script>
	import { flyabsolute } from '$lib/animations';
	import LoginWebauthn from '$lib/login/login-webauthn.svelte';
	import { quartInOut } from 'svelte/easing';
</script>

<div
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
>
	<LoginWebauthn />
</div>
