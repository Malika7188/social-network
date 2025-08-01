import { useState, useEffect } from 'react';
import { useGroupService } from '@/services/groupService';
import styles from '@/styles/GroupEvents.module.css';
import { BASE_URL } from '@/utils/constants';
import CreateEventModal from '@/components/events/CreateEventModal';

const GroupEvents = ({ groupId }) => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [searchTerm, setSearchTerm] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const [events, setEvents] = useState([]);
    const [loading, setLoading] = useState(true);
    const eventsPerPage = 10;

    const { getGroupEvents, respondToEvent, createEvent } = useGroupService();

    useEffect(() => {
        fetchEvents();
    }, [groupId]);

    const fetchEvents = async () => {
        try {
            setLoading(true);
            const fetchedEvents = await getGroupEvents(groupId);
            
            // Ensure events are properly formatted
            const formattedEvents = Array.isArray(fetchedEvents) ? fetchedEvents.map(event => ({
                id: event.ID || event.id,
                title: event.Title || event.title,
                eventDate: event.EventDate || event.eventDate,
                bannerPath: event.BannerPath?.String || event.bannerPath,
                goingCount: event.GoingCount || event.goingCount || 0,
                userResponse: event.UserResponse || event.userResponse,
                description: event.Description || event.description
            })) : [];
            
            setEvents(formattedEvents);
        } catch (error) {
            console.error('Error fetching events:', error);
            setEvents([]);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateEvent = async (eventData) => {
        try {
            await createEvent(groupId, eventData);
            setIsModalOpen(false);
            fetchEvents(); // Refresh events list
        } catch (error) {
            console.error('Error creating event:', error);
        }
    };

    const handleEventResponse = async (eventId, response) => {
        try {
            await respondToEvent(eventId, response);
            fetchEvents(); // Refresh events to update attendance
        } catch (error) {
            console.error('Error responding to event:', error);
        }
    };

    // Filter and pagination logic
    const filteredEvents = events?.filter(event =>
        event?.title?.toLowerCase().includes(searchTerm.toLowerCase())
    ) || [];

    const totalPages = Math.ceil(filteredEvents.length / eventsPerPage);
    const currentEvents = filteredEvents.slice(
        (currentPage - 1) * eventsPerPage,
        currentPage * eventsPerPage
    );

    if (loading) {
        return <div className={styles.loading}>Loading events...</div>;
    }

    return (
        <div className={styles.eventsContainer}>
            <div className={styles.header}>
                <div className={styles.headerLeft}>
                    <h2>Group Events</h2>
                    <button 
                        className={styles.createButton}
                        onClick={() => setIsModalOpen(true)}
                    >
                        <i className="fas fa-plus"></i>
                        Create Event
                    </button>
                </div>
                <div className={styles.searchBar}>
                    <i className="fas fa-search"></i>
                    <input
                        type="text"
                        placeholder="Search events..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
            </div>

            <div className={styles.eventsGrid}>
                {currentEvents.length === 0 ? (
                    <div className={styles.noEvents}>
                        <p>No events found</p>
                    </div>
                ) : (
                    currentEvents.map(event => (
                        <div key={event.id} className={styles.eventCard}>
                            <div className={styles.eventBanner}>
                                <img 
                                    src={event.bannerPath ? 
                                        `${BASE_URL}/uploads/${event.bannerPath}` : 
                                        '/default-event-banner.jpg'} 
                                    alt={event.title} 
                                />
                            </div>
                            <div className={styles.eventInfo}>
                                <h3>{event.title}</h3>
                                <p className={styles.description}>{event.description}</p>
                                <p className={styles.date}>
                                    <i className="fas fa-calendar"></i>
                                    {new Date(event.eventDate).toLocaleDateString('en-US', {
                                        month: 'long',
                                        day: 'numeric',
                                        year: 'numeric',
                                        hour: '2-digit',
                                        minute: '2-digit'
                                    })}
                                </p>
                                <p className={styles.attendees}>
                                    <i className="fas fa-users"></i>
                                    {event.goingCount || 0} going
                                </p>
                                <div className={styles.actions}>
                                     {event.userResponse ? (
                                        // User has already responded
                                        <div className={styles.responseStatus}>
                                            <span className={styles.currentResponse}>
                                                {event.userResponse === 'going' ? 
                                                    <><i className="fas fa-check-circle"></i> Going</> : 
                                                    <><i className="fas fa-times-circle"></i> Not Going</>
                                                }
                                            </span>
                                            <button 
                                                className={styles.changeResponse}
                                                onClick={() => handleEventResponse(event.id, 
                                                    event.userResponse === 'going' ? 'not_going' : 'going'
                                                )}
                                            >
                                                Change mind?
                                            </button>
                                        </div>
                                    ) : (
                                        // User hasn't responded yet
                                        <div className={styles.responseOptions}>
                                            <button 
                                                className={styles.goingButton}
                                                onClick={() => handleEventResponse(event.id, 'going')}
                                            >
                                                <i className="fas fa-check"></i> Going?
                                            </button>
                                            <button 
                                                className={styles.notGoingButton}
                                                onClick={() => handleEventResponse(event.id, 'not_going')}
                                            >
                                                <i className="fas fa-times"></i> Not Going?
                                            </button>
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    ))
                )}
            </div>

            {totalPages > 1 && (
                <div className={styles.pagination}>
                    {Array.from({ length: totalPages }, (_, i) => (
                        <button
                            key={i + 1}
                            onClick={() => setCurrentPage(i + 1)}
                            className={currentPage === i + 1 ? styles.active : ''}
                        >
                            {i + 1}
                        </button>
                    ))}
                </div>
            )}

            <CreateEventModal 
                groupId={groupId}
                isOpen={isModalOpen} 
                onClose={() => setIsModalOpen(false)}
                onSubmit={fetchEvents}
            />
        </div>
    );
};

export default GroupEvents;