<script lang="ts">
	import Button from '$lib/Button.svelte'
	import type { ShopItem } from '$lib/shop'
	import Icon from '$lib/icon/Icon.svelte'
	import ImageUpload from './ImageUpload.svelte'

	interface Props {
		product: ShopItem
		updateProduct: () => void
	}
	let { product = $bindable(), updateProduct }: Props = $props()

	const addVariant = () => {
		product.variants = [
			...product.variants,
			{
				type: '',
				options: [
					// Automatically create blanks for the admin to fill in.
					{ text: '', additionalPrice: 0 },
					{ text: '', additionalPrice: 0 }
				]
			}
		]
	}
	const removeVariant = (i: number) => () => {
		product.variants.splice(i, 1)
		product.variants = product.variants
	}

	const addVariantOption = (i: number) => () => {
		product.variants[i].options = [
			...product.variants[i].options,
			{
				text: '',
				additionalPrice: 0
			}
		]
	}
	const removeVariantOption = (i: number, j: number) => () => {
		product.variants[i].options.splice(j, 1)
		product.variants[i].options = product.variants[i].options
	}

	const addImageURL = () => {
		product.imageURLs = [
			...product.imageURLs,
			{
				selectedOptions: [],
				url: ''
			}
		]
	}
	const removeImageURL = (i: number) => () => {
		product.imageURLs.splice(i, 1)
		product.imageURLs = product.imageURLs
	}
</script>

<div class="config">
	<span>ID</span>
	<input bind:value={product.id} placeholder="Product ID (auto-generated)" disabled />
	<span>Sale Period</span>
	<input bind:value={product.salePeriod} placeholder="Product ID (auto-generated)" disabled />
	<span>Name</span>
	<input bind:value={product.name} placeholder="Product Name" />
	<span>Enabled</span>
	<input bind:checked={product.enabled} type="checkbox" />
	<span>Base Price (cents)</span>
	<input type="number" bind:value={product.basePrice} />
	<span>Default Image URL</span>
	<ImageUpload bind:value={product.defaultImageURL} />
	<span class="header">Variants</span>
	{#each product.variants as variant, i}
		<div class="flex gap-x-2 self-start">
			<input bind:value={product.variants[i].type} placeholder="Variant Name" />
			<button onclick={removeVariant(i)}><Icon name="trash" class="size-4" /></button>
		</div>
		<div class="variant-config">
			<div class="col-span-4 flex items-center gap-1">
				<span>Size Chart URL</span>
				<ImageUpload bind:value={product.variants[i].chart_url} raw />
			</div>
			{#each variant.options as _, j}
				<input bind:value={product.variants[i].options[j].text} placeholder="Option Name" />
				<span>Additional Price</span>
				<input type="number" bind:value={product.variants[i].options[j].additionalPrice} />
				<button onclick={removeVariantOption(i, j)}><Icon name="trash" class="size-4" /></button>
			{/each}
			<Button size="md" onClick={addVariantOption(i)}>Add {variant.type} Option</Button>
		</div>
	{/each}
	<span class="col-span-2 flex"><Button size="md" onClick={addVariant}>Add Variant</Button></span>
	<span class="header">Variant Image URLs</span>
	{#each product.imageURLs as _, i}
		<div class="flex items-center justify-between self-start">
			<span>Variant {i + 1}</span>
			<button onclick={removeImageURL(i)}><Icon name="trash" class="size-4" /></button>
		</div>
		<div class="flex flex-col justify-start gap-1">
			<div class="flex gap-1">
				{#each product.variants as variant, j}
					<select bind:value={product.imageURLs[i].selectedOptions[j]}>
						<option value={null}>ANY {variant.type.toUpperCase()}</option>
						{#each variant.options as option}
							<option value={option.text}>{option.text}</option>
						{/each}
					</select>
				{/each}
			</div>
			<ImageUpload bind:value={product.imageURLs[i].url} />
		</div>
	{/each}
	<span class="col-span-2 flex">
		<Button size="md" onClick={addImageURL}>Add Image URL Variant</Button>
	</span>
	<div class="flex"><Button onClick={updateProduct}>Update Product</Button></div>
</div>

<style lang="postcss">
	.config {
		@apply grid items-center gap-x-2 gap-y-1;
		grid-template-columns: auto 1fr;
	}
	input {
		@apply max-w-64 border border-black px-1;
	}
	input:disabled {
		@apply cursor-not-allowed border-gray-600 bg-gray-200 text-gray-500;
	}
	input[type='checkbox'] {
		@apply justify-self-start;
		max-width: unset;
	}
	.header {
		@apply col-span-2 text-xl;
	}

	.variant-config {
		@apply grid w-fit items-center gap-x-1 gap-y-1;
		grid-template-columns: auto auto auto auto;
	}
	select {
		@apply border border-black bg-white px-1;
	}
</style>
