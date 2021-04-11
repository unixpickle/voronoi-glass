# voronoi-glass

Have you ever looked through a shower door made of intentionally uneven glass? Everything looks distorted, but maybe also beautiful. Now, with this silly program, you can make any picture look this way.

# Example

First, here is an example of me, distorted with this program:

<img src="example/alex.jpg" width="250">

Next, we apply the program with different settings to this base image:

<img src="example/landscape.jpg" width="250">

There are two noising options. The default manipulates each vertex in a mesh, while the latter manipulates full voronoi cells at once. To enable the latter behavior, use `-use-nn`. Running the script with `-use-nn` produces this output:

<img src="example/landscape_nn_noise05.jpg" width="250">

There are a number of parameters which control how the refraction works. For now, we will only vary the `-noise` argument, which determines how uneven the glass is. With `-noise 0.5` (the default), we get this:

<img src="example/landscape_noise05.jpg" width="250">

With less noise (0.2), we get a smoother version:

<img src="example/landscape_noise02.jpg" width="250">

If we increase the noise to 1.0, we start to see large changes at almost every polygon:

<img src="example/landscape_noise1.jpg" width="250">

# How it works

## Nearest neighbor randomness

In this method, we use refraction on an imaginary "plane" with normals determined randomly for each Voronoi cell. We generate a grid of random points, and a random normal for each point. We cast rays onto a plane towards the image, and use a refraction algorithm to bend the ray according to the normal for the nearest neighbor of each ray collision. We then march the ray a fixed Z distance (thus translating X and Y), and use the pixel from the original image at this point.

## Vertex-level randomness

First, we generate a Voronoi diagram because they provide a natural way to break up a plane into polygons. I use a very inefficient O(N^3) algorithm for this, since it was very simple to implement; there exist better algorithms to do this in O(N log(N)) complexity. The resulting diagram looks like so (the mesh is in red, and the black dots are the original points used to generate it):

![Voronoi diagram](example/voronoi.png)

Next, we convert this diagram into a 3D mesh, randomly perturbing the Z coordinate of each vertex to get an uneven terrain. The amount of random noise is a controllable parameter. I'll note that the change to a given vertex is also controlled by the shape of its surrounding triangles. Generally, sliver triangles will have tips that can safely move a great deal, while the two close points can hardly be moved without greatly changing the normal. The program measures this "sensitivity" and acts accordingly.

Now that we have a rough mesh representing our "glass", we can cast rays onto it. When a ray hits the glass, we apply the refraction formula `sin(theta1) = sin(theta)*n1/n2`. We then take the refracted (bent) ray and carry it forward for a certain fixed Z distance, at which point it "hits" the original image. We also reflect the image in all directions to prevent any scenarios where a ray is refracted off of the image.
