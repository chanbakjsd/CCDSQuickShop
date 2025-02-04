<script lang="ts">
	import { formatPrice } from '$lib/cart';

	export let name: string;
	export let imageURL: string;
	export let basePrice: number;
	export let selected = false;
</script>

<button on:click class:selected class="border-box flex flex-col shadow-lg hover:shadow-xl">
	<img class="size-48" alt={`Picture of ${name}`} src={imageURL} />
	<div class="px-4 py-2 text-left">
		<p class="text-xl">{name}</p>
		<p class="text-sm">S$ {formatPrice(basePrice)}</p>
	</div>
</button>

<style lang="postcss">
	.selected {
		box-shadow: 0 0 2px 3px #0f2b50;
	}
</style>
