package src.pas.chess.moveorder;

// SYSTEM IMPORTS
import edu.bu.chess.search.DFSTreeNode;
import edu.bu.chess.game.move.Move;
import edu.bu.chess.game.move.MoveType;
import edu.bu.chess.game.move.MovementMove;
import edu.bu.chess.game.piece.PieceType;
import edu.bu.chess.utils.Coordinate;

import java.util.LinkedList;
import java.util.List;

// JAVA PROJECT IMPORTS
import src.pas.chess.moveorder.DefaultMoveOrderer;

public class CustomMoveOrderer
    extends Object
{

	/**
	 * TODO: implement me!
	 * This method should perform move ordering. Remember, move ordering is how alpha-beta pruning gets part of its power from.
	 * You want to see nodes which are beneficial FIRST so you can prune as much as possible during the search (i.e. be faster)
	 * @param nodes. The nodes to order (these are children of a DFSTreeNode) that we are about to consider in the search.
	 * @return The ordered nodes.
	 */
	public static List<DFSTreeNode> order(List<DFSTreeNode> nodes) {
        List<DFSTreeNode> captureNodes = new LinkedList<>();
        List<DFSTreeNode> promotionNodes = new LinkedList<>();
        List<DFSTreeNode> castlingNodes = new LinkedList<>();
        List<DFSTreeNode> enPassantNodes = new LinkedList<>();
        List<DFSTreeNode> highValueThreatNodes = new LinkedList<>();
        List<DFSTreeNode> otherNodes = new LinkedList<>();

        for (DFSTreeNode node : nodes) {
            Move move = node.getMove();
            if (move != null) {
                switch (move.getType()) {
                    case CAPTUREMOVE:
                        captureNodes.add(node);
                        break;
                    case PROMOTEPAWNMOVE: // Promote pawns (usually a good move)
                        promotionNodes.add(node);
                        break;
                    case CASTLEMOVE: // Castling (important for king safety)
                        castlingNodes.add(node);
                        break;
                    case ENPASSANTMOVE: // En passant capture
                        enPassantNodes.add(node);
                        break;
                    default:
                        if (threatensHighValuePiece(node)) {
                            highValueThreatNodes.add(node);
                        } else {
                            otherNodes.add(node);
                        }
                        break;
                }
            } else {
                otherNodes.add(node); // Add moves that don’t fit the criteria
            }
        }

        // Merge all prioritized lists: capture > promotion > castling > en passant > threats > others
        captureNodes.addAll(promotionNodes);
        captureNodes.addAll(castlingNodes);
        captureNodes.addAll(enPassantNodes);
        captureNodes.addAll(highValueThreatNodes);
        captureNodes.addAll(otherNodes);

        return captureNodes;
    }

    /**
     * Checks if the move threatens a high-value piece.
     * A "high-value" piece is typically a Queen, Rook, or Bishop.
     *
     * @param node. The node to check if it threatens a high-value piece.
     * @return true if the move threatens a high-value piece, false otherwise.
     */
    private static boolean threatensHighValuePiece(DFSTreeNode node) {
        Move move = node.getMove();
        if (move instanceof MovementMove) {  // Check if the move is a MovementMove
            MovementMove movementMove = (MovementMove) move;
            Coordinate targetPosition = movementMove.getTargetPosition();  // Use getTargetPosition()
            PieceType targetType = node.getGame().getBoard().getPieceAtPosition(targetPosition).getType();

            // Consider pieces like Queen, Rook, or Bishop as high-value pieces.
            return targetType == PieceType.QUEEN || targetType == PieceType.ROOK || targetType == PieceType.BISHOP;
        }
        return false;
    }

}




