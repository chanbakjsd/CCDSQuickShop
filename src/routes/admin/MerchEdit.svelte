<script lang="ts">
	import { onMount } from 'svelte';

	import { fetchProducts, updateProduct } from '$lib/api';
	import { emptyShopItem, type ShopItem } from '$lib/shop';
	import MerchList from '$lib/MerchList.svelte';
	import ProductEdit from './ProductEdit.svelte';
	import ErrorBoundary from '$lib/ErrorBoundary.svelte';

	let loading = $state(true);
	let items: ShopItem[] = $state([]);
	let fetchError: unknown = $state();
	onMount(() => {
		fetchProducts(true)
			.then((x) => {
				loading = false;
				items = x;
			})
			.catch((e) => {
				fetchError = e;
			});
	});

	$effect(() => {
		if (!loading && (items.length === 0 || items[items.length - 1].id !== '')) {
			items.push(emptyShopItem('Add New Product'));
		}
	});

	const variants = $derived(Array(items.length).fill({}));
	let selectedItemIdx = $state(-1);
	let selectedItem: ShopItem | undefined = $state(undefined);

	$effect(() => {
		selectedItem = selectedItemIdx >= 0 ? items[selectedItemIdx] : undefined;
	});

	let updateError: unknown = $state();
	const update = async () => {
		if (!selectedItem) return;
		try {
			const updatedItem = await updateProduct(selectedItem);
			items[selectedItemIdx] = updatedItem;
		} catch (e) {
			updateError = e;
		}
	};
</script>

<div class="flex flex-col gap-4 p-4">
	<ErrorBoundary error={fetchError}>
		<MerchList {items} {variants} bind:value={selectedItemIdx} />
		{#if selectedItem}
			<hr />
			<ProductEdit bind:product={selectedItem} updateProduct={update} />
			<ErrorBoundary error={updateError} />
		{/if}
	</ErrorBoundary>
</div>
