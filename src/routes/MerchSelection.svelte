<script lang="ts">
	import { fade, fly } from 'svelte/transition'
	import { formatPrice, type CartItem } from '$lib/cart'
	import { resolveImageURL, tentativePrice, toArrayVariant, type ShopItem } from '$lib/shop'
	import Options from '$lib/Options.svelte'
	import Button from '$lib/Button.svelte'
	import MerchList from '$lib/MerchList.svelte'
	import ZoomableImage from '$lib/ZoomableImage.svelte'

	interface Props {
		items: ShopItem[]
		addItem: (item: CartItem) => any
	}

	const { items, addItem }: Props = $props()
	// All the variants that the user has committed to (clicked on).
	const chosenVariants: Record<string, string | undefined>[] = []
	$effect(() => {
		while (chosenVariants.length < items.length) chosenVariants.push({})
	})
	// The index of the merch being selected right now.
	let activeMerch = $state(-1)
	const selectedItem = $derived(activeMerch >= 0 ? items[activeMerch] : undefined)

	// The cart item representing the user's current choice.
	const cartItem = $derived.by(() => {
		if (!selectedItem) return undefined
		const activeVariant = chosenVariants[activeMerch]
		const arrayVariant = toArrayVariant(selectedItem, activeVariant)
		// Everything must be selected.
		if (arrayVariant.some((x) => !x)) return undefined
		return {
			id: selectedItem.id,
			name: selectedItem.name,
			variant: arrayVariant.map((x, i) => ({ type: selectedItem.variants[i].type, option: x! })),
			imageURL: resolveImageURL(selectedItem, activeVariant),
			unitPrice: tentativePrice(selectedItem, activeVariant),
			amount: 1
		}
	})

	const selectedVariant = $derived(activeMerch >= 0 ? chosenVariants[activeMerch] : {})
	const updatePreview = (type: string) => (variant?: string) =>
		(chosenVariants[activeMerch][type] = variant)

	const previewImage = $derived(selectedItem ? resolveImageURL(selectedItem, selectedVariant) : '')
	const tryAddItem = () => {
		if (!cartItem) return
		addItem({ ...cartItem })
	}
</script>

<div class="flex flex-col gap-4">
	<MerchList {items} variants={chosenVariants} bind:value={activeMerch} />
	{#if selectedItem}
		<hr class="border-gray-300" transition:fade />
		<div
			class="grid grid-cols-1 justify-between gap-2 lg:grid-cols-[1fr,auto]"
			transition:fly={{ y: 100 }}
		>
			<div class="row-start-2 flex min-w-0 flex-col gap-2 break-words lg:row-start-1">
				<p class="text-3xl">{selectedItem.name}</p>
				<p class="text-xl">
					S$ {formatPrice(tentativePrice(selectedItem, selectedVariant) / 100)}
				</p>
				<div class="options">
					{#each selectedItem.variants as variant}
						<span class="py-1 transition-all">{variant.type}</span>
						<Options
							options={variant.options}
							bind:value={chosenVariants[activeMerch][variant.type]}
							updatePreview={updatePreview(variant.type)}
						/>
					{/each}
				</div>
				<div class="mt-2">
					<Button onClick={tryAddItem} disabled={!cartItem}>Add to Cart</Button>
				</div>
			</div>
			<div class="row-start-1">
				<ZoomableImage
					imageURL={previewImage}
					class="size-48 flex-shrink-0"
					name={selectedItem.name}
				/>
			</div>
		</div>
	{/if}
</div>

<style lang="postcss">
	.options {
		@apply grid gap-x-4 gap-y-2;
		grid-template-columns: auto 1fr;
	}
</style>
