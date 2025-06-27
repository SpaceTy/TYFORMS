<template>
  <div class="starry-background">
    <div class="star-container"
         :style="{
           transform: `perspective(1000px) 
                      rotateX(${rotation.x}deg) 
                      rotateY(${rotation.y}deg)`
         }">
      <div v-for="star in stars" 
           :key="star.id" 
           class="star"
           :style="{
             transform: `translate3d(${star.x}px, ${star.y}px, ${star.z}px)`,
             opacity: star.opacity,
             width: star.size + 'px',
             height: star.size + 'px'
           }">
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'StarryBackground',
  data() {
    return {
      stars: [],
      rotation: {
        x: 0,
        y: 0
      },
      targetRotation: {
        x: 0,
        y: 0
      },
      spaceSize: {
        width: 0,
        height: 0,
        depth: 2000 // Z-depth of our space
      }
    }
  },
  mounted() {
    window.addEventListener('mousemove', this.handleMouseMove)
    window.addEventListener('resize', this.handleResize)
    this.handleResize()
    this.updateRotation()
  },
  beforeUnmount() {
    window.removeEventListener('mousemove', this.handleMouseMove)
    window.removeEventListener('resize', this.handleResize)
  },
  methods: {
    generateStars() {
      const count = 300
      const stars = []
      
      for (let i = 0; i < count; i++) {
        const x = (Math.random() * this.spaceSize.width) - (this.spaceSize.width / 2)
        const y = (Math.random() * this.spaceSize.height) - (this.spaceSize.height / 2)
        const z = (Math.random() * -this.spaceSize.depth)
        
        const zRatio = Math.abs(z / this.spaceSize.depth)
        const size = 1 + (2 * (1 - zRatio))
        const opacity = 0.2 + (0.8 * (1 - zRatio))
        
        stars.push({
          id: i,
          x,
          y,
          z,
          size,
          opacity
        })
      }
      
      return stars
    },
    updateRotation() {
      // Smooth interpolation towards target rotation
      this.rotation.x += (this.targetRotation.x - this.rotation.x) * 0.1
      this.rotation.y += (this.targetRotation.y - this.rotation.y) * 0.1
      
      // Emit the current rotation
      this.$emit('rotation-change', { ...this.rotation })
      
      requestAnimationFrame(this.updateRotation)
    },
    handleMouseMove(event) {
      const { clientX, clientY } = event
      const { innerWidth, innerHeight } = window
      
      this.targetRotation.y = ((clientX / innerWidth) - 0.5) * 40
      this.targetRotation.x = ((clientY / innerHeight) - 0.5) * -40
    },
    handleResize() {
      this.spaceSize.width = window.innerWidth * 1.5
      this.spaceSize.height = window.innerHeight * 1.5
      this.stars = this.generateStars()
    }
  }
}
</script>

<style scoped>
.starry-background {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: black;
  z-index: 0;
  overflow: hidden;
}

.star-container {
  position: absolute;
  width: 100%;
  height: 100%;
  transform-style: preserve-3d;
  transform-origin: center center;
}

.star {
  position: absolute;
  left: 50%;
  top: 50%;
  background-color: white;
  border-radius: 50%;
  transform-style: preserve-3d;
  pointer-events: none;
}
</style> 