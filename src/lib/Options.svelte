<script lang="ts">
	import { formatPrice } from './cart';

	export let options: {
		text: string;
		additionalPrice?: number;
	}[];
	export let value: string | undefined = undefined;
	export let previewValue: string | undefined = undefined;

	let currentHover: string | undefined;
	const select = (option: string) => () => {
		if (value === option) {
			value = undefined;
			return;
		}
		value = option;
	};
	const hover = (option: string) => () => {
		currentHover = option;
	};
	const unhover = () => {
		currentHover = undefined;
	};

	$: currentOffset = options.find((x) => x.text === value)?.additionalPrice ?? 0;
	$: offsetPriceDisplay = options
		.map((x) => (x.additionalPrice ?? 0) - currentOffset)
		.map((x) => {
			if (x > 0) return `(+$${formatPrice(x / 100)})`;
			if (x < 0) return `(-$${formatPrice(-x / 100)})`;
			return '';
		});
	$: {
		if (currentHover) {
			previewValue = currentHover;
		} else {
			previewValue = value;
		}
	}
</script>

<div class="flex flex-wrap gap-x-2 gap-y-1" on:pointerleave={unhover}>
	{#each options as option, i}
		<button
			class:selected={value === option.text}
			on:click={select(option.text)}
			on:pointerenter={hover(option.text)}
		>
			{option.text}
			{offsetPriceDisplay[i]}
		</button>
	{/each}
</div>

<style lang="postcss">
	button {
		@apply min-w-8 bg-gray-200 px-2 py-1 transition-all hover:bg-gray-300;
	}
	.selected {
		@apply bg-white hover:bg-white;
		box-shadow: 0 0 1px 2px #0f2b50;
	}
</style>
