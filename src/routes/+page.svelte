<script lang="ts">
	import type { CartItem, Coupon } from '$lib/cart';
	import type { ShopItem } from '$lib/shop';
	import Cart from './Cart.svelte';
	import MerchSelection from './MerchSelection.svelte';

	const items: ShopItem[] = [
		{
			id: 'sticker',
			name: 'Stickers',
			basePrice: 199,
			variants: [
				{
					type: 'Image',
					options: [
						{
							text: 'SCDS Club Logo (Light)',
							additionalPrice: 0
						},
						{
							text: 'SCDS Club Logo (Dark)',
							additionalPrice: 0
						},
						{
							text: 'Eat Sleep Code Repeat',
							additionalPrice: 100
						}
					]
				}
			],
			defaultImageURL: 'https://placehold.co/400',
			imageURLs: [
				{ selectedOptions: ['SCDS Club Logo (Dark)'], url: 'https://placehold.co/200' },
				{ selectedOptions: ['Eat Sleep Code Repeat'], url: 'https://placehold.co/300' }
			]
		},
		{
			id: 'tshirt-dino',
			name: 'Dino T-Shirt',
			basePrice: 1699,
			variants: [
				{
					type: 'Material',
					options: [
						{ text: 'Dri-fit', additionalPrice: 400 },
						{ text: 'Cotton', additionalPrice: 0 }
					]
				},
				{
					type: 'Color',
					options: [
						{ text: 'Navy', additionalPrice: 0 },
						{ text: 'White', additionalPrice: 0 },
						{ text: 'Black', additionalPrice: 0 }
					]
				},
				{
					type: 'Size',
					options: [
						{ text: '3XS', additionalPrice: 0 },
						{ text: '2XS', additionalPrice: 0 },
						{ text: 'XS', additionalPrice: 0 },
						{ text: 'S', additionalPrice: 0 },
						{ text: 'M', additionalPrice: 0 },
						{ text: 'L', additionalPrice: 0 },
						{ text: 'XL', additionalPrice: 0 },
						{ text: '2XL', additionalPrice: 0 },
						{ text: '3XL', additionalPrice: 0 }
					]
				}
			],
			defaultImageURL: 'https://placehold.co/500',
			imageURLs: [
				{ selectedOptions: [undefined, 'White'], url: 'https://placehold.co/200' },
				{ selectedOptions: [undefined, 'Black'], url: 'https://placehold.co/300' }
			]
		}
	];

	let cart: CartItem[] = [];
	const addToCart = (item: CartItem) => {
		// Try to search for existing entries and add amount instead.
		for (let i = 0; i < cart.length; i++) {
			if (
				cart[i].id !== item.id ||
				cart[i].variant.length !== item.variant.length ||
				cart[i].amount + item.amount > 50
			) {
				continue;
			}
			const variantMatch = cart[i].variant.every((x) =>
				item.variant.some((y) => x.type === y.type && x.option === y.option)
			);
			if (!variantMatch) {
				continue;
			}
			// This is a match! Just add amount.
			cart[i].amount += item.amount;
			return;
		}
		cart = [...cart, item];
	};
	const coupon: Coupon[] = [
		{
			requirements: [{ type: 'purchase_count', amount: 3 }],
			couponCode: 'BUYTOGETHER3',
			discount: {
				type: 'percentage',
				amount: 5
			}
		},
		{
			requirements: [{ type: 'purchase_count', amount: 5 }],
			couponCode: 'BUYTOGETHER5',
			discount: {
				type: 'percentage',
				amount: 10
			}
		}
	];
</script>

<svelte:head><title>SCDS Merch Store</title></svelte:head>

<main class="grid grid-cols-1 md:grid-cols-3">
	<div class="left-panel">
		<div
			class="sticky top-0 -my-2 flex items-center justify-center gap-2 bg-white pb-2 md:justify-start"
		>
			<img
				src="https://ntuscds.com/scse-logo/scds-logo.png"
				class="size-16"
				alt="Logo for SCDS Club"
			/>
			<p class="text-2xl font-bold text-brand">Merch Store</p>
		</div>
		<div class="px-2 md:py-2">
			<MerchSelection {items} addItem={addToCart} />
		</div>
	</div>
	<div class="right-panel"><Cart bind:cart availableCoupons={coupon} /></div>
</main>

<style lang="postcss">
	.left-panel {
		@apply col-span-2 mx-2 h-fit border-b border-gray-400 pb-4 pt-2;
		@apply flex flex-col gap-4;
		@apply md:mx-0 md:my-2 md:border-b-0 md:border-r md:px-4 md:py-0;
		@apply md:h-[calc(100vh-1rem)] md:overflow-y-auto;
	}
	.right-panel {
		@apply overflow-y-auto px-4 py-4;
		@apply md:h-screen;
	}
</style>
