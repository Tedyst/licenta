import { cubicOut } from 'svelte/easing';
import { fly, type TransitionConfig } from 'svelte/transition';

export function flyabsolute(
	node: Element,
	{
		delay = 0,
		duration = 400,
		easing = cubicOut,
		x = 0,
		y = 0,
		opacity = 0,
		otherStyling = ''
	} = {}
): TransitionConfig {
	const flyConfig = fly(node, { delay, duration, easing, x, y, opacity });
	return {
		...flyConfig,
		css: (t, u) =>
			`opacity: ${
				t * u
			}; ${otherStyling} position: absolute; margin-left: 0; margin-right: 0; left: 0; right: 0; ${flyConfig?.css?.(
				t,
				u
			)};`
	};
}
