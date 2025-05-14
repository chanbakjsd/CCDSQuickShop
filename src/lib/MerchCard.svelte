<script lang="ts">
	import { formatPrice } from '$lib/cart'

	interface Props {
		name: string
		imageURL: string
		basePrice: number
		selected?: boolean
		onclick?: () => any
	}

	const { name, imageURL, basePrice, selected = false, onclick }: Props = $props()
</script>

<button
	onclick={() => onclick?.()}
	class:selected
	class="border-box flex flex-col shadow-lg hover:shadow-xl"
>
	<img class="size-48" alt={`Picture of ${name}`} src={imageURL} />
	<div class="flex max-w-48 flex-grow flex-col justify-between break-words px-4 py-2 text-left">
		<p class="text-xl">{name}</p>
		<p class="text-sm">S$ {formatPrice(basePrice)}</p>
	</div>
</button>

<style lang="postcss">
	.selected {
		box-shadow: 0 0 2px 3px #0f2b50;
	}
</style>
