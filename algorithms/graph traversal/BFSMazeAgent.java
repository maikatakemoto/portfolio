package src.labs.stealth.agents;

// SYSTEM IMPORTS
import edu.bu.labs.stealth.agents.MazeAgent;
import edu.bu.labs.stealth.graph.Vertex;
import edu.bu.labs.stealth.graph.Path;


import edu.cwru.sepia.environment.model.state.State.StateView;


import java.util.HashSet;       // will need for bfs
import java.util.Queue;         // will need for bfs
import java.util.LinkedList;    // will need for bfs
import java.util.Set;           // will need for bfs


// JAVA PROJECT IMPORTS


public class BFSMazeAgent
    extends MazeAgent
{

    public BFSMazeAgent(int playerNum)
    {
        super(playerNum);
    }

    @Override
    public Path search(Vertex src,
                       Vertex goal,
                       StateView state)
    {
        // Initialize data structures --> keep track of paths to goal AND visited vertices
        Queue<Path> myQ = new LinkedList<>();      
        Set<Vertex> visited = new HashSet<>();      
 
        // Add first "path" and mark starting node as visited
        Path startPath = new Path(src);            
        myQ.add(startPath);                         
        visited.add(src);                           

        while (!myQ.isEmpty()) {
            Path currPath = myQ.poll();                         // Remove path object from the front of queue 
            Vertex currVertex = currPath.getDestination();      // Get the vertex of the current path 

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
                    if (!visited.contains(neighbor)) {
                        if (!state.isResourceAt(neighborX, neighborY) && (state.inBounds(neighborX, neighborY))) {
                            if (neighbor.equals(goal)) {
                                return currPath;
                            } else {
                                Path neighborPath = new Path(neighbor, 1.0f, currPath); 
                                visited.add(neighbor);
                                myQ.add(neighborPath);
                            }
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

