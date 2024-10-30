package src.labs.stealth.agents;

// SYSTEM IMPORTS
import edu.bu.labs.stealth.agents.MazeAgent;
import edu.bu.labs.stealth.graph.Vertex;
import edu.bu.labs.stealth.graph.Path;


import edu.cwru.sepia.environment.model.state.State.StateView;
import edu.cwru.sepia.util.Direction;                           // Directions in Sepia


import java.util.Comparator;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.PriorityQueue; // heap in java
import java.util.Set;


// JAVA PROJECT IMPORTS


public class DijkstraMazeAgent
    extends MazeAgent
{

    public DijkstraMazeAgent(int playerNum)
    {
        super(playerNum);
    }

    @Override
    public Path search(Vertex src,
                       Vertex goal,
                       StateView state)
    {
        // Initialize data structures 
        // Priority queue is ordered by smallest to largest cost
        PriorityQueue<Path> myPQ = new PriorityQueue<>((p1, p2) -> Float.compare(p1.getTrueCost(), p2.getTrueCost()));
        Set<Vertex> visited = new HashSet<>();     

        // Cost of directions 
        float Up = 10f;
        float Down = 1f;
        float Side = 5f;
  
        // Add first "path" and mark starting node as visited
        Path startPath = new Path(src);            
        myPQ.add(startPath);                         
        visited.add(src);                           

        while (!myPQ.isEmpty()) {
            Path currPath = myPQ.poll();                       // Pop path object from the top of stack
            Vertex currVertex = currPath.getDestination();     // Get the vertex of the current path 

            int currX = currVertex.getXCoordinate();
            int currY = currVertex.getYCoordinate();

            // Traverse through neighbors 
            for (int i =- 1; i <= 1; i++) {
                for (int j =- 1; j <= 1; j++) {
                    if (i == 0 && j == 0) {
                        // No need to explore current vertex
                        continue;                           
                    }
                    int neighborX = currX + i;
                    int neighborY = currY + j;
                    Vertex neighbor = new Vertex(neighborX, neighborY);

                    if (neighbor.equals(goal)) {
                        return currPath;
                    }
         
                    float directionCost = 0f;

                    if (i != 0 && j != 0) {
                        if (j < 0) {
                            directionCost = (float) Math.sqrt(Math.pow(Up, 2) + Math.pow(Side, 2));
                        } else {
                            directionCost = (float) Math.sqrt(Math.pow(Down, 2) + Math.pow(Side, 2));
                        }
                    } else if (j != 0) {
                        if (j < 0) {
                            directionCost = 10f;
                        } else {
                            directionCost = 1f;
                        }
                    } else if (i != 0) {
                        directionCost = 5f;
                    }

                    if (!visited.contains(neighbor)) {
                        if (!state.isResourceAt(neighborX, neighborY) && (state.inBounds(neighborX, neighborY))) {
                            Path neighborPath = new Path(neighbor, directionCost, currPath); 
                            visited.add(neighbor);
                            myPQ.add(neighborPath);
                        }
                    }
                }
            }
        }
        return null;
    }

    @Override
    public boolean shouldReplacePlan(StateView state)
    {
        return false;
    }

}
