import boxGreyIcon from '@/assets/images/icons/box-grey.svg'
import makeupBlackIcon from '@/assets/images/icons/Makeup-black.svg'
import smartwatchImg from '@/assets/images/thumbnails/smartwatch.png'
import bagBlackIcon from '@/assets/images/icons/bag-black.svg'
import milkBlackIcon from '@/assets/images/icons/milk-black.svg'
import glassBlackIcon from '@/assets/images/icons/glass-black.svg'
import galleryDefaultIcon from '@/assets/images/icons/gallery-default.svg'

export { galleryDefaultIcon, boxGreyIcon }

export function formatCurrency(amount) {
  return new Intl.NumberFormat('id-ID').format(Number(amount) || 0)
}

export function formatNumber(number) {
  return new Intl.NumberFormat('id-ID').format(Number(number) || 0)
}

export function getCategoryIcon(categoryName) {
  if (!categoryName) return boxGreyIcon
  const categoryIcons = {
    cosmetics: makeupBlackIcon,
    makeup: makeupBlackIcon,
    electronics: smartwatchImg,
    fashion: bagBlackIcon,
    food: milkBlackIcon,
    beverages: glassBlackIcon
  }
  return categoryIcons[categoryName.toLowerCase()] || boxGreyIcon
}
