ArrayList<Ball> balls;
PGraphics pg;
PImage img;

void setup() {
  size(600, 600, P3D);
  float fov = PI/3.0; // Field of view
  float cameraZ = 1000; // Adjust this value to move the camera backward
  perspective(fov, float(width)/float(height), cameraZ/10.0, cameraZ*10.0);
  camera(width/2.0, height/2.0, cameraZ, width/2.0, height/2.0, 0.0, 0.0, 1.0, 0.0);

  pg = createGraphics(width, height, P3D);

  balls = new ArrayList<Ball>();
}

void draw() {
  // Create an off-screen buffer for drawing the 3D space
  pg.beginDraw();
  pg.background(0);
  pg.translate(width / 2, height / 2, 0);
  pg.noFill();
  pg.stroke(255);
  box(width);
  pg.endDraw();

  // Render the off-screen buffer
  image(pg, 0, 0);
  background(0); 
  
  // 3D Space
  translate(width / 2, height / 2, 0);
  noFill();
  stroke(255);
  box(width);

  
  // Create balls
  for (Ball ball : balls) {
    ball.update();
    ball.updateRotation();
    ball.checkBoundaryCollision();
    ball.decayEnergy();
    ball.display();
  }
    
  // Check for collision between balls
  for (int i = 0; i < balls.size(); i++) {
    for (int j = i + 1; j < balls.size(); j++) {
      balls.get(i).checkCollision(balls.get(j));
    }
  }
}

// Shoot a ball at position of mouse
void mouseClicked() {
  img = loadImage(sketchPath("data/texture")+Integer.toString((int)random(1,5))+".jpg");
  PVector initial_position = new PVector(mouseX, mouseY, 0);
  PVector initial_speed = new PVector(random(-5, 5), random(-5, 5), random(5));
  PVector rotation_speed = new PVector(random(0.02, 0.5) * PI, random(0.02, 0.5) * PI, random(0.02, 0.5) * PI);
  int radius = 50;
  smooth();
  Ball ball = new Ball(initial_position, initial_speed, rotation_speed, radius, img);
  balls.add(ball);
}

class Ball {
  PVector position;
  PVector velocity;
  PVector rotation;
  PImage texture;
  PShape globe;
  float radius;
  float mass;
  boolean onGround;
  float rotationFriction;
  
  // Create sphere and manually add speed, position, rotation, etc...
  Ball(PVector initial_position, PVector initial_speed, PVector rotation_speed, int radius, PImage img) {
    this.position = initial_position.copy();
    this.velocity = initial_speed.copy();
    this.rotation = rotation_speed.copy();
    this.radius = radius;
    this.mass = radius * 0.1;
    
    rotationFriction = 0.995; 
    
    // Adding textures to spheres
    this.texture = img;
    noStroke();
    noFill();
    globe = createShape(SPHERE, radius);
    globe.setTexture(img);
  }
  
  void update() {
    position.add(velocity); 
    rotation.mult(rotationFriction);
  }
  
  void updateRotation() {
    float rotationSpeedFactor = 0.5; 
    rotation.add(0, rotationSpeedFactor, 0);
  }
  
  void checkBoundaryCollision() {
    float cubeSize = 600; 
    if (position.x > cubeSize / 2 - radius) {
      position.x = cubeSize / 2 - radius;
      velocity.x *= -1;
    } else if (position.x < -cubeSize / 2 + radius) {
      position.x = -cubeSize / 2 + radius;
      velocity.x *= -1;
    }
    
    if (position.y > cubeSize / 2 - radius) {
      position.y = cubeSize / 2 - radius;
      velocity.y *= -1;
    } else if (position.y < -cubeSize / 2 + radius) {
      position.y = -cubeSize / 2 + radius;
      velocity.y *= -1;
    }
    
    if (position.z > cubeSize / 2 - radius) {
      position.z = cubeSize / 2 - radius;
      velocity.z *= -1;
    } else if (position.z < -cubeSize / 2 + radius) {
      position.z = -cubeSize / 2 + radius;
      velocity.z *= -1;
    }
  }

  void checkCollision(Ball other) {

    // Get distances between balls 
    PVector distanceVect = PVector.sub(other.position, position);

    // Calculate magnitude of the vector separating balls
    float distanceVectMag = distanceVect.mag();

    // Min dist before they touch
    float minDistance = radius + other.radius;

    if (distanceVectMag < minDistance) {
      float distanceCorrection = (minDistance-distanceVectMag)/2.0;
      PVector d = distanceVect.copy();
      PVector correctionVector = d.normalize().mult(distanceCorrection);
      other.position.add(correctionVector);
      position.sub(correctionVector);

      // Find angle of distanceVect
      float theta  = distanceVect.heading();
      float sine = sin(theta);
      float cosine = cos(theta);

      PVector[] bTemp = {
        new PVector(), new PVector()
      };

      bTemp[1].x  = cosine * distanceVect.x + sine * distanceVect.y;
      bTemp[1].y  = cosine * distanceVect.y - sine * distanceVect.x;

      // Rotate temporary velocities
      PVector[] vTemp = {
        new PVector(), new PVector()
      };

      vTemp[0].x  = cosine * velocity.x + sine * velocity.y;
      vTemp[0].y  = cosine * velocity.y - sine * velocity.x;
      vTemp[1].x  = cosine * other.velocity.x + sine * other.velocity.y;
      vTemp[1].y  = cosine * other.velocity.y - sine * other.velocity.x;

      PVector[] vFinal = {  
        new PVector(), new PVector()
      };

      // Final rotated velocity for b[0]
      vFinal[0].x = ((mass - other.mass) * vTemp[0].x + 2 * other.mass * vTemp[1].x) / (mass + other.mass);
      vFinal[0].y = vTemp[0].y;

      // Final rotated velocity for b[0]
      vFinal[1].x = ((other.mass - mass) * vTemp[1].x + 2 * mass * vTemp[0].x) / (mass + other.mass);
      vFinal[1].y = vTemp[1].y;

      // Avoid clumping
      bTemp[0].x += vFinal[0].x;
      bTemp[1].x += vFinal[1].x;

      // Rotate balls
      PVector[] bFinal = { 
        new PVector(), new PVector()
      };

      bFinal[0].x = cosine * bTemp[0].x - sine * bTemp[0].y;
      bFinal[0].y = cosine * bTemp[0].y + sine * bTemp[0].x;
      bFinal[1].x = cosine * bTemp[1].x - sine * bTemp[1].y;
      bFinal[1].y = cosine * bTemp[1].y + sine * bTemp[1].x;

      // Update balls to screen position
      other.position.x = position.x + bFinal[1].x;
      other.position.y = position.y + bFinal[1].y;

      position.add(bFinal[0]);

      // Update velocities
      velocity.x = cosine * vFinal[0].x - sine * vFinal[0].y;
      velocity.y = cosine * vFinal[0].y + sine * vFinal[0].x;
      other.velocity.x = cosine * vFinal[1].x - sine * vFinal[1].y;
      other.velocity.y = cosine * vFinal[1].y + sine * vFinal[1].x;
      
      // Add friction when the balls collide
      float friction = 0.95;
      velocity.mult(friction);
      other.velocity.mult(friction);
    }
  }
  
  void decayEnergy() {
    PVector gravity = new PVector(0, 0.1, 0);
    float energyDecay = 0.995;
    
    velocity.add(gravity);
    velocity.mult(energyDecay);
  }

  void display() {
    pushMatrix();
    translate(position.x, position.y, position.z);
    rotateX(rotation.x);
    rotateY(rotation.y);
    rotateZ(rotation.z);
    texture(texture);
    shape(globe);
    popMatrix();
  }
}
 
